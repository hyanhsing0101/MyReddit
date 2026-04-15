package controller

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"myreddit/pkg/snowflake"
	"myreddit/settings"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var errUploadPublicBase = errors.New("upload public_url not configured")

// sniffImageExt 仅依据文件头识别 PNG/JPEG/GIF/WebP，不信任客户端 Content-Type。
func sniffImageExt(data []byte) (ext string, ok bool) {
	if len(data) < 12 {
		return "", false
	}
	if len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return ".png", true
	}
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return ".jpg", true
	}
	if len(data) >= 6 && (bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a"))) {
		return ".gif", true
	}
	if len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return ".webp", true
	}
	return "", false
}

func uploadMaxBytes() int64 {
	if settings.Conf.Upload == nil || settings.Conf.Upload.MaxBytes <= 0 {
		return 5 << 20 // 5 MiB
	}
	return settings.Conf.Upload.MaxBytes
}

func uploadPublicBase() (string, error) {
	if settings.Conf.Upload == nil {
		return "", errUploadPublicBase
	}
	base := strings.TrimSpace(settings.Conf.Upload.PublicURL)
	if base == "" {
		return "", errUploadPublicBase
	}
	return strings.TrimRight(base, "/"), nil
}

func requestPublicBase(c *gin.Context) string {
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	if p := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto")); p != "" {
		scheme = p
	}
	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		host = "127.0.0.1:8081"
	}
	return fmt.Sprintf("%s://%s", scheme, host)
}

// UploadImageHandler 登录用户上传图片（multipart 字段名 file），返回可插入 Markdown 的绝对 URL。
func UploadImageHandler(c *gin.Context) {
	if settings.Conf.Upload == nil || !settings.Conf.Upload.Enabled || strings.TrimSpace(settings.Conf.Upload.Dir) == "" {
		ResponseError(c, CodeUploadDisabled)
		return
	}
	publicBase, err := uploadPublicBase()
	if err != nil {
		// 开发态兜底：未配 public_url 时，按当前请求 Host 组装 URL，避免本地调试被配置阻塞。
		if settings.Conf.Mode == "dev" {
			publicBase = requestPublicBase(c)
		} else {
			zap.L().Warn("upload disabled: public_url empty")
			ResponseError(c, CodeUploadDisabled)
			return
		}
	}

	dir, err := filepath.Abs(settings.Conf.Upload.Dir)
	if err != nil {
		zap.L().Error("upload dir abs", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	maxB := uploadMaxBytes()
	if err := c.Request.ParseMultipartForm(maxB + (1 << 20)); err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	var fh *multipart.FileHeader
	if fhs := c.Request.MultipartForm.File["file"]; len(fhs) > 0 {
		fh = fhs[0]
	}
	if fh == nil {
		ResponseErrorWithMsg(c, CodeInvalidParam, "missing form field file")
		return
	}
	if fh.Size > maxB {
		ResponseError(c, CodeUploadTooLarge)
		return
	}

	src, err := fh.Open()
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	defer src.Close()

	limited := io.LimitReader(src, maxB+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		zap.L().Error("upload read", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	if int64(len(data)) > maxB {
		ResponseError(c, CodeUploadTooLarge)
		return
	}
	ext, ok := sniffImageExt(data)
	if !ok {
		ResponseError(c, CodeUploadInvalidImage)
		return
	}

	name := strconv.FormatInt(snowflake.GenID(), 10) + ext
	finalPath := filepath.Join(dir, name)
	tmp, err := os.CreateTemp(dir, ".part-")
	if err != nil {
		zap.L().Error("upload createtemp", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	tmpPath := tmp.Name()
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr != nil || cerr != nil {
		_ = os.Remove(tmpPath)
		zap.L().Error("upload write", zap.Error(werr), zap.NamedError("close", cerr))
		ResponseError(c, CodeServerBusy)
		return
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		zap.L().Error("upload rename", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	url := fmt.Sprintf("%s/uploads/%s", publicBase, name)
	ResponseSuccess(c, gin.H{"url": url})
}

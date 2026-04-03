const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://127.0.0.1:8081";

export const API_SUCCESS_CODE = 1000;

export type ApiResponse<T> = {
  code: number;
  msg: unknown;
  data: T;
};

async function parseJson<T>(res: Response): Promise<ApiResponse<T>> {
  const text = await res.text();
  try {
    return JSON.parse(text) as ApiResponse<T>;
  } catch {
    throw new Error(`无效响应 (${res.status}): ${text.slice(0, 200)}`);
  }
}

export function apiErrorMessage(body: ApiResponse<unknown>): string {
  const { msg } = body;
  if (typeof msg === "string") return msg;
  if (msg !== null && typeof msg === "object") {
    return JSON.stringify(msg);
  }
  return `错误码 ${body.code}`;
}

export async function apiSignup(payload: {
  username: string;
  password: string;
  re_password: string;
}): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/signup`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiLogin(payload: {
  username: string;
  password: string;
}): Promise<
  ApiResponse<{ access_token: string; refresh_token: string } | null>
> {
  const res = await fetch(`${API_BASE}/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return parseJson(res);
}

export async function apiCreatePost(
  accessToken: string,
  payload: { title: string; content: string },
): Promise<ApiResponse<null>> {
  const res = await fetch(`${API_BASE}/post`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(payload),
  });
  return parseJson<null>(res);
}

export async function apiPing(accessToken: string): Promise<string> {
  const res = await fetch(`${API_BASE}/ping`, {
    method: "GET",
    headers: { Authorization: `Bearer ${accessToken}` },
  });
  const ct = res.headers.get("content-type") ?? "";
  if (ct.includes("application/json")) {
    const body = (await res.json()) as ApiResponse<unknown>;
    throw new Error(apiErrorMessage(body));
  }
  return res.text();
}

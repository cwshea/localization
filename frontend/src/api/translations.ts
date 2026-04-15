import type {
  SourceString,
  Translation,
  CreateRequest,
  RetranslateRequest,
} from "../types";

const API_BASE = "/api";

async function request<T>(
  path: string,
  options?: RequestInit
): Promise<T> {
  const controller = new AbortController();
  const timeout = setTimeout(() => controller.abort(), 90_000);

  let res: Response;
  try {
    res = await fetch(`${API_BASE}${path}`, {
      headers: { "Content-Type": "application/json" },
      signal: controller.signal,
      ...options,
    });
  } catch (err) {
    clearTimeout(timeout);
    if (err instanceof DOMException && err.name === "AbortError") {
      throw new Error("Request timed out. The translation may still be processing.");
    }
    throw err;
  }
  clearTimeout(timeout);

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

function normalizeSource(s: SourceString): SourceString {
  return { ...s, translations: s.translations ?? [] };
}

export async function listSourceStrings(): Promise<SourceString[]> {
  const data = await request<SourceString[]>("/source-strings");
  return data.map(normalizeSource);
}

export async function getSourceString(id: string): Promise<SourceString> {
  return normalizeSource(await request<SourceString>(`/source-strings/${id}`));
}

export async function createSourceString(
  data: CreateRequest
): Promise<SourceString> {
  return normalizeSource(
    await request<SourceString>("/source-strings", {
      method: "POST",
      body: JSON.stringify(data),
    })
  );
}

export function updateSourceString(
  id: string,
  text: string
): Promise<SourceString> {
  return request<SourceString>(`/source-strings/${id}`, {
    method: "PUT",
    body: JSON.stringify({ text }),
  });
}

export function deleteSourceString(id: string): Promise<void> {
  return request<void>(`/source-strings/${id}`, { method: "DELETE" });
}

export function updateTranslation(
  id: string,
  translatedText: string
): Promise<Translation> {
  return request<Translation>(`/translations/${id}`, {
    method: "PUT",
    body: JSON.stringify({ translated_text: translatedText }),
  });
}

export function deleteTranslation(id: string): Promise<void> {
  return request<void>(`/translations/${id}`, { method: "DELETE" });
}

export async function retranslate(
  sourceId: string,
  data: RetranslateRequest
): Promise<SourceString> {
  return normalizeSource(
    await request<SourceString>(`/source-strings/${sourceId}/retranslate`, {
      method: "POST",
      body: JSON.stringify(data),
    })
  );
}

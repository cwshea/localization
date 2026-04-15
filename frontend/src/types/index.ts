export interface Translation {
  id: string;
  source_id: string;
  locale: string;
  translated_text: string;
  llm_provider: string;
  translated_at: string;
  updated_at: string;
}

export interface SourceString {
  id: string;
  text: string;
  created_at: string;
  updated_at: string;
  translations: Translation[];
  warning?: string;
}

export interface CreateRequest {
  text: string;
  locales: string[];
  llm_providers: string[];
}

export interface RetranslateRequest {
  locales: string[];
  llm_providers: string[];
}

export const LOCALES: Record<string, string> = {
  "en-GB": "British English",
  es: "Spanish",
  "zh-Hant": "Traditional Chinese",
  "zh-Hans": "Simplified Chinese",
  hi: "Hindi",
};

export const LLM_PROVIDERS: Record<string, string> = {
  chatgpt5: "ChatGPT 5",
  gemini25: "Gemini 2.5",
};

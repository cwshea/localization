import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { createSourceString } from "../api/translations";
import LocaleCheckboxes from "./LocaleCheckboxes";
import LlmSelector from "./LlmSelector";

export default function TranslationForm() {
  const navigate = useNavigate();
  const [text, setText] = useState("");
  const [locales, setLocales] = useState<string[]>([]);
  const [providers, setProviders] = useState<string[]>(["chatgpt5"]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");

    if (!text.trim()) {
      setError("Source text is required");
      return;
    }
    if (locales.length === 0) {
      setError("Select at least one target language");
      return;
    }
    if (providers.length === 0) {
      setError("Select at least one LLM provider");
      return;
    }

    setLoading(true);
    try {
      const result = await createSourceString({
        text: text.trim(),
        locales,
        llm_providers: providers,
      });
      if (result.warning) {
        // Source was created but translations failed -- navigate with warning
        navigate(`/source/${result.id}`, {
          state: { warning: result.warning },
        });
      } else {
        navigate(`/source/${result.id}`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create translation");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="page">
      <h2>New Translation</h2>
      <form onSubmit={handleSubmit} className="form">
        {error && <div className="error-message">{error}</div>}

        <div className="form-group">
          <label className="field-label" htmlFor="source-text">
            Source Text (American English)
          </label>
          <textarea
            id="source-text"
            value={text}
            onChange={(e) => setText(e.target.value)}
            rows={4}
            placeholder="Enter text to translate..."
            disabled={loading}
          />
        </div>

        <LocaleCheckboxes selected={locales} onChange={setLocales} />
        <LlmSelector selected={providers} onChange={setProviders} />

        <div className="form-actions">
          <button
            type="button"
            className="btn btn-secondary"
            onClick={() => navigate("/")}
            disabled={loading}
          >
            Cancel
          </button>
          <button type="submit" className="btn btn-primary" disabled={loading}>
            {loading ? "Translating..." : "Translate"}
          </button>
        </div>
      </form>
    </div>
  );
}

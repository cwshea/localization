import { useEffect, useState } from "react";
import { useParams, useNavigate, useLocation } from "react-router-dom";
import {
  getSourceString,
  updateSourceString,
  updateTranslation,
  deleteTranslation,
  deleteSourceString,
  retranslate,
} from "../api/translations";
import type { SourceString } from "../types";
import { LOCALES, LLM_PROVIDERS } from "../types";
import ConfirmDelete from "./ConfirmDelete";
import LocaleCheckboxes from "./LocaleCheckboxes";
import LlmSelector from "./LlmSelector";

export default function TranslationDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const location = useLocation();

  const [source, setSource] = useState<SourceString | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [warning, setWarning] = useState(
    (location.state as { warning?: string })?.warning || ""
  );

  // Source text editing
  const [editingSource, setEditingSource] = useState(false);
  const [sourceText, setSourceText] = useState("");
  const [savingSource, setSavingSource] = useState(false);

  // Translation editing
  const [editingTranslation, setEditingTranslation] = useState<string | null>(null);
  const [editedText, setEditedText] = useState("");

  // Delete confirmation
  const [deleteTarget, setDeleteTarget] = useState<{ type: "source" | "translation"; id: string } | null>(null);

  // Retranslate
  const [showRetranslate, setShowRetranslate] = useState(false);
  const [retranslateLocales, setRetranslateLocales] = useState<string[]>([]);
  const [retranslateProviders, setRetranslateProviders] = useState<string[]>(["chatgpt5"]);
  const [retranslating, setRetranslating] = useState(false);

  const fetchData = async () => {
    if (!id) return;
    try {
      const data = await getSourceString(id);
      setSource(data);
      setSourceText(data.text);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [id]);

  const handleSaveSource = async () => {
    if (!id || !sourceText.trim()) return;
    setSavingSource(true);
    try {
      const updated = await updateSourceString(id, sourceText.trim());
      setSource((prev) => (prev ? { ...prev, text: updated.text, updated_at: updated.updated_at } : prev));
      setEditingSource(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update");
    } finally {
      setSavingSource(false);
    }
  };

  const handleSaveTranslation = async (translationId: string) => {
    if (!editedText.trim()) return;
    try {
      const updated = await updateTranslation(translationId, editedText.trim());
      setSource((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          translations: prev.translations.map((t) =>
            t.id === translationId ? updated : t
          ),
        };
      });
      setEditingTranslation(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to update");
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      if (deleteTarget.type === "source") {
        await deleteSourceString(deleteTarget.id);
        navigate("/");
      } else {
        await deleteTranslation(deleteTarget.id);
        setSource((prev) => {
          if (!prev) return prev;
          return {
            ...prev,
            translations: prev.translations.filter((t) => t.id !== deleteTarget.id),
          };
        });
      }
      setDeleteTarget(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete");
    }
  };

  const handleRetranslate = async () => {
    if (!id || retranslateLocales.length === 0 || retranslateProviders.length === 0) return;
    setRetranslating(true);
    setWarning("");
    try {
      const updated = await retranslate(id, {
        locales: retranslateLocales,
        llm_providers: retranslateProviders,
      });
      setSource(updated);
      setShowRetranslate(false);
      setRetranslateLocales([]);
      if (updated.warning) {
        setWarning(updated.warning);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to retranslate");
    } finally {
      setRetranslating(false);
    }
  };

  if (loading) return <div className="page"><p>Loading...</p></div>;
  if (!source) return <div className="page"><p>Source string not found.</p></div>;

  return (
    <div className="page">
      <button className="btn btn-secondary btn-back" onClick={() => navigate("/")}>
        &larr; Back
      </button>

      {error && <div className="error-message">{error}</div>}
      {warning && (
        <div className="warning-message">
          Translation warning: {warning}
          <button className="warning-dismiss" onClick={() => setWarning("")}>&times;</button>
        </div>
      )}

      <div className="detail-card">
        <div className="detail-header">
          <h2>Source Text</h2>
          <div className="detail-actions">
            {!editingSource && (
              <button className="btn btn-secondary btn-sm" onClick={() => setEditingSource(true)}>
                Edit
              </button>
            )}
            <button
              className="btn btn-danger btn-sm"
              onClick={() => setDeleteTarget({ type: "source", id: source.id })}
            >
              Delete All
            </button>
          </div>
        </div>

        {editingSource ? (
          <div className="edit-area">
            <textarea
              value={sourceText}
              onChange={(e) => setSourceText(e.target.value)}
              rows={3}
              disabled={savingSource}
            />
            <div className="edit-actions">
              <button
                className="btn btn-secondary btn-sm"
                onClick={() => {
                  setEditingSource(false);
                  setSourceText(source.text);
                }}
                disabled={savingSource}
              >
                Cancel
              </button>
              <button className="btn btn-primary btn-sm" onClick={handleSaveSource} disabled={savingSource}>
                {savingSource ? "Saving..." : "Save"}
              </button>
            </div>
          </div>
        ) : (
          <p className="source-text">{source.text}</p>
        )}

        <div className="meta">
          Created: {new Date(source.created_at).toLocaleString()}
          {" | "}
          Updated: {new Date(source.updated_at).toLocaleString()}
        </div>
      </div>

      <div className="translations-section">
        <div className="section-header">
          <h3>Translations ({source.translations?.length ?? 0})</h3>
          <button className="btn btn-primary btn-sm" onClick={() => setShowRetranslate(!showRetranslate)}>
            {showRetranslate ? "Cancel" : "+ Add / Retranslate"}
          </button>
        </div>

        {showRetranslate && (
          <div className="retranslate-form">
            <LocaleCheckboxes selected={retranslateLocales} onChange={setRetranslateLocales} />
            <LlmSelector selected={retranslateProviders} onChange={setRetranslateProviders} />
            <button
              className="btn btn-primary"
              onClick={handleRetranslate}
              disabled={retranslating || retranslateLocales.length === 0 || retranslateProviders.length === 0}
            >
              {retranslating ? "Translating..." : "Translate"}
            </button>
          </div>
        )}

        {!source.translations?.length ? (
          <p className="empty-state">No translations yet.</p>
        ) : (
          <div className="locale-groups">
            {Object.entries(
              source.translations.reduce<Record<string, typeof source.translations>>((acc, t) => {
                (acc[t.locale] ??= []).push(t);
                return acc;
              }, {})
            ).map(([locale, translations]) => (
              <div key={locale} className="locale-group">
                <div className="locale-group-header">
                  <span className="locale-badge">{LOCALES[locale] || locale}</span>
                </div>
                <div className="locale-group-cards">
                  {translations.map((t) => (
                    <div key={t.id} className="translation-card">
                      <div className="translation-card-header">
                        <span className="provider-badge">{LLM_PROVIDERS[t.llm_provider] || t.llm_provider}</span>
                        <div className="translation-card-actions">
                          {editingTranslation !== t.id && (
                            <button
                              className="btn btn-secondary btn-sm"
                              onClick={() => {
                                setEditingTranslation(t.id);
                                setEditedText(t.translated_text);
                              }}
                            >
                              Edit
                            </button>
                          )}
                          <button
                            className="btn btn-danger btn-sm"
                            onClick={() => setDeleteTarget({ type: "translation", id: t.id })}
                          >
                            Delete
                          </button>
                        </div>
                      </div>

                      {editingTranslation === t.id ? (
                        <div className="edit-area">
                          <textarea
                            value={editedText}
                            onChange={(e) => setEditedText(e.target.value)}
                            rows={3}
                          />
                          <div className="edit-actions">
                            <button className="btn btn-secondary btn-sm" onClick={() => setEditingTranslation(null)}>
                              Cancel
                            </button>
                            <button className="btn btn-primary btn-sm" onClick={() => handleSaveTranslation(t.id)}>
                              Save
                            </button>
                          </div>
                        </div>
                      ) : (
                        <p className="translated-text">{t.translated_text}</p>
                      )}

                      <div className="meta">
                        Translated: {new Date(t.translated_at).toLocaleString()}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {deleteTarget && (
        <ConfirmDelete
          message={
            deleteTarget.type === "source"
              ? "Delete this source string and ALL its translations?"
              : "Delete this translation?"
          }
          onConfirm={handleDelete}
          onCancel={() => setDeleteTarget(null)}
        />
      )}
    </div>
  );
}

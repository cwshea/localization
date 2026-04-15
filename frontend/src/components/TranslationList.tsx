import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { listSourceStrings, deleteSourceString } from "../api/translations";
import type { SourceString } from "../types";
import { LOCALES, LLM_PROVIDERS } from "../types";
import ConfirmDelete from "./ConfirmDelete";

type ViewMode = "detail" | "summary";

export default function TranslationList() {
  const [sources, setSources] = useState<SourceString[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [deleteTarget, setDeleteTarget] = useState<SourceString | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>(
    () => (localStorage.getItem("localization-view-mode") as ViewMode) || "detail"
  );

  const toggleView = (mode: ViewMode) => {
    setViewMode(mode);
    localStorage.setItem("localization-view-mode", mode);
  };

  const fetchData = async () => {
    try {
      const data = await listSourceStrings();
      setSources(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to load");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const handleDelete = async () => {
    if (!deleteTarget) return;
    try {
      await deleteSourceString(deleteTarget.id);
      setSources((prev) => prev.filter((s) => s.id !== deleteTarget.id));
      setDeleteTarget(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete");
    }
  };

  if (loading) return <div className="page"><p>Loading...</p></div>;

  return (
    <div className="page">
      <div className="page-header">
        <h2>Translations</h2>
        <div className="page-header-actions">
          <div className="view-toggle">
            <button
              className={`view-toggle-btn ${viewMode === "summary" ? "active" : ""}`}
              onClick={() => toggleView("summary")}
              title="Summary view"
            >
              Summary
            </button>
            <button
              className={`view-toggle-btn ${viewMode === "detail" ? "active" : ""}`}
              onClick={() => toggleView("detail")}
              title="Detail view"
            >
              Detail
            </button>
          </div>
          <Link to="/new" className="btn btn-primary">
            + New Translation
          </Link>
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}

      {sources.length === 0 ? (
        <p className="empty-state">
          No translations yet. Create one to get started.
        </p>
      ) : viewMode === "summary" ? (
        <div className="table-container">
          <table>
            <thead>
              <tr>
                <th>Source Text</th>
                <th>Languages</th>
                <th>Created</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {sources.map((source) => (
                <tr key={source.id}>
                  <td className="source-text-cell">
                    <Link to={`/source/${source.id}`}>
                      {source.text.length > 80
                        ? source.text.slice(0, 80) + "..."
                        : source.text}
                    </Link>
                  </td>
                  <td>
                    <div className="locale-badges">
                      {[...new Set(source.translations?.map((t) => t.locale))].map((locale) => (
                        <span key={locale} className="locale-badge">
                          {LOCALES[locale] || locale}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="timestamp-cell">
                    {new Date(source.created_at).toLocaleDateString()}
                  </td>
                  <td>
                    <div className="row-actions">
                      <Link to={`/source/${source.id}`} className="btn btn-secondary btn-sm">
                        Edit
                      </Link>
                      <button
                        className="btn btn-danger btn-sm"
                        onClick={() => setDeleteTarget(source)}
                      >
                        Delete
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      ) : (
        <div className="source-cards">
          {sources.map((source) => (
            <div key={source.id} className="source-card">
              <div className="source-card-header">
                <div className="source-card-text">
                  <Link to={`/source/${source.id}`} className="source-link">
                    {source.text}
                  </Link>
                  <span className="meta">
                    Created {new Date(source.created_at).toLocaleDateString()}
                  </span>
                </div>
                <div className="source-card-actions">
                  <Link to={`/source/${source.id}`} className="btn btn-secondary btn-sm">
                    Edit
                  </Link>
                  <button
                    className="btn btn-danger btn-sm"
                    onClick={() => setDeleteTarget(source)}
                  >
                    Delete
                  </button>
                </div>
              </div>

              {source.translations?.length > 0 && (
                <table className="translations-table">
                  <thead>
                    <tr>
                      <th>Locale</th>
                      <th>Translated Text</th>
                      <th>LLM</th>
                      <th>Translated At</th>
                    </tr>
                  </thead>
                  <tbody>
                    {source.translations.map((t) => (
                      <tr key={t.id}>
                        <td>
                          <span className="locale-badge">
                            {LOCALES[t.locale] || t.locale}
                          </span>
                        </td>
                        <td className="translated-text-cell">{t.translated_text}</td>
                        <td>
                          <span className="provider-badge">
                            {LLM_PROVIDERS[t.llm_provider] || t.llm_provider}
                          </span>
                        </td>
                        <td className="timestamp-cell">
                          {new Date(t.translated_at).toLocaleString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              )}

              {(!source.translations || source.translations.length === 0) && (
                <p className="no-translations">No translations yet</p>
              )}
            </div>
          ))}
        </div>
      )}

      {deleteTarget && (
        <ConfirmDelete
          message={`Delete "${deleteTarget.text.slice(0, 50)}..." and all its translations?`}
          onConfirm={handleDelete}
          onCancel={() => setDeleteTarget(null)}
        />
      )}
    </div>
  );
}

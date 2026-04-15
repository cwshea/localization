import { LLM_PROVIDERS } from "../types";

interface Props {
  selected: string[];
  onChange: (providers: string[]) => void;
}

export default function LlmSelector({ selected, onChange }: Props) {
  const toggle = (value: string) => {
    if (selected.includes(value)) {
      onChange(selected.filter((v) => v !== value));
    } else {
      onChange([...selected, value]);
    }
  };

  return (
    <div className="llm-selector">
      <label className="field-label">LLM Providers</label>
      <div className="checkbox-group">
        {Object.entries(LLM_PROVIDERS).map(([value, label]) => (
          <label key={value} className="checkbox-label">
            <input
              type="checkbox"
              value={value}
              checked={selected.includes(value)}
              onChange={() => toggle(value)}
            />
            {label}
          </label>
        ))}
      </div>
    </div>
  );
}

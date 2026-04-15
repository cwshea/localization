import { LOCALES } from "../types";

interface Props {
  selected: string[];
  onChange: (locales: string[]) => void;
}

export default function LocaleCheckboxes({ selected, onChange }: Props) {
  const toggle = (locale: string) => {
    if (selected.includes(locale)) {
      onChange(selected.filter((l) => l !== locale));
    } else {
      onChange([...selected, locale]);
    }
  };

  return (
    <div className="locale-checkboxes">
      <label className="field-label">Target Languages</label>
      <div className="checkbox-group">
        {Object.entries(LOCALES).map(([code, name]) => (
          <label key={code} className="checkbox-label">
            <input
              type="checkbox"
              checked={selected.includes(code)}
              onChange={() => toggle(code)}
            />
            {name} ({code})
          </label>
        ))}
      </div>
    </div>
  );
}

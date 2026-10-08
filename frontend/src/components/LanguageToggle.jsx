const options = [
  { value: 'pt-BR', label: 'PT-BR' },
  { value: 'en', label: 'EN' }
];

function LanguageToggle({ language, label, onChange }) {
  return (
    <div className="st-seg" role="group" aria-label={label}>
      {options.map((option) => (
        <button
          key={option.value}
          type="button"
          className="st-seg__item"
          aria-pressed={language === option.value}
          onClick={() => onChange(option.value)}
        >
          {option.label}
        </button>
      ))}
    </div>
  );
}

export default LanguageToggle;

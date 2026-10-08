// Caminho longo quebra linha em vez de ser abreviado; o title mantém o
// caminho completo para quem passa o mouse (design system, PathField; #74).
function PathField({ label, path, emptyText, actionLabel, onAction, disabled }) {
  return (
    <div className="st-path">
      <span className="st-path__label">
        <i />
        {label}
      </span>
      {path ? (
        <span className="st-path__value" title={path}>
          {path}
        </span>
      ) : (
        <span className="st-path__value st-path__value--empty">{emptyText}</span>
      )}
      {actionLabel && (
        <button
          type="button"
          className="st-link st-path__action"
          onClick={onAction}
          disabled={disabled}
        >
          {actionLabel}
        </button>
      )}
    </div>
  );
}

export default PathField;

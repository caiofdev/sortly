function Switch({ checked, label, disabled, onChange }) {
  return (
    <button
      type="button"
      role="switch"
      className="st-switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      onClick={() => onChange(!checked)}
    />
  );
}

export default Switch;

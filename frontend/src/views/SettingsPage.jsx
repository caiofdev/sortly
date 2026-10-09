import SegmentedControl from '../components/SegmentedControl';
import Switch from '../components/Switch';

// Rótulo e dica de cada critério: byDuration → settingsByDuration e
// settingsHintByDuration. A ordem e a trava do último ligado vêm do backend (#44).
const suffix = (key) => `${key[0].toUpperCase()}${key.slice(1)}`;

function SettingsPage({ labels, theme, criteria, onThemeChange, onCriterionChange }) {
  const themes = [
    { value: 'dark', label: labels.themeDark },
    { value: 'light', label: labels.themeLight }
  ];

  return (
    <div className="st-card st-settings">
      <div className="st-settings__row">
        <div>
          <p className="st-settings__title">{labels.settingsTheme}</p>
          <p className="st-settings__hint">{labels.settingsHintTheme}</p>
        </div>
        <SegmentedControl
          label={labels.settingsTheme}
          options={themes}
          value={theme}
          onChange={onThemeChange}
        />
      </div>

      {criteria.map((item) => {
        const title = labels[`settings${suffix(item.key)}`];
        return (
          <div key={item.key} className="st-settings__row">
            <div>
              <p className="st-settings__title">{title}</p>
              <p className="st-settings__hint">
                {item.locked
                  ? labels.settingsLockedHint
                  : labels[`settingsHint${suffix(item.key)}`]}
              </p>
            </div>
            <Switch
              checked={item.enabled}
              label={title}
              disabled={item.locked}
              onChange={(enabled) => onCriterionChange(item.key, enabled)}
            />
          </div>
        );
      })}
    </div>
  );
}

export default SettingsPage;

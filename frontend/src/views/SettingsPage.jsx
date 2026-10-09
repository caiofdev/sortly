import SegmentedControl from '../components/SegmentedControl';
import Switch from '../components/Switch';

// Rótulo e dica de cada critério: byDuration → settingsByDuration e
// settingsHintByDuration. A ordem e a trava do último ligado vêm do backend (#44).
const suffix = (key) => `${key[0].toUpperCase()}${key.slice(1)}`;

function SettingsPage({
  labels,
  theme,
  duplicates,
  criteria,
  onThemeChange,
  onDuplicatesChange,
  onCriterionChange
}) {
  const themes = [
    { value: 'dark', label: labels.themeDark },
    { value: 'light', label: labels.themeLight }
  ];
  const duplicateOptions = [
    { value: 'rename', label: labels.duplicatesRename },
    { value: 'skip', label: labels.duplicatesSkip },
    { value: 'replace', label: labels.duplicatesReplace }
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

      <div className="st-settings__row">
        <div>
          <p className="st-settings__title">{labels.settingsDuplicates}</p>
          <p className="st-settings__hint">{labels.settingsHintDuplicates}</p>
        </div>
        <SegmentedControl
          label={labels.settingsDuplicates}
          options={duplicateOptions}
          value={duplicates}
          onChange={onDuplicatesChange}
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

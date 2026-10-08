import Switch from '../components/Switch';

// Rótulo e dica de cada critério: byDuration → settingsByDuration e
// settingsHintByDuration. A ordem e a trava do último ligado vêm do backend (#44).
const suffix = (key) => `${key[0].toUpperCase()}${key.slice(1)}`;

function SettingsPage({ labels, criteria, onCriterionChange }) {
  return (
    <section className="st-card st-settings" aria-label={labels.settingsCriteriaTitle}>
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
    </section>
  );
}

export default SettingsPage;

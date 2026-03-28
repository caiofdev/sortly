import appLogo from '../assets/app-logo.svg';
import LanguageToggle from './LanguageToggle';
import OrganizerSettingsPanel from './OrganizerSettingsPanel';

function OrganizerHeader({ language, onLanguageChange, subtitle, labels, settingsOpen, onToggleSettings, options, onOptionChange }) {
  return (
    <header className="space-y-3 text-center">
      <div className="flex items-center justify-between gap-3 pb-2">
        <LanguageToggle language={language} onChange={onLanguageChange} />
        <OrganizerSettingsPanel
          isOpen={settingsOpen}
          onToggle={onToggleSettings}
          labels={labels}
          options={options}
          onOptionChange={onOptionChange}
        />
      </div>

      <div className="flex items-center justify-center gap-3">
        <img src={appLogo} alt="Logo do app" className="h-14 w-14" />
        <h1 className="text-3xl font-bold tracking-tight md:text-4xl">Sortly</h1>
      </div>

      <p className="text-sm text-[#94A3B8] md:text-base">{subtitle}</p>
    </header>
  );
}

export default OrganizerHeader;

import logo from '../assets/sortly-logo.png';
import { HistoryIcon, OrganizeIcon, SettingsIcon } from './icons';
import LanguageToggle from './LanguageToggle';

export const PAGES = [
  { id: 'organize', labelKey: 'navOrganize', Icon: OrganizeIcon },
  { id: 'history', labelKey: 'navHistory', Icon: HistoryIcon },
  { id: 'settings', labelKey: 'navSettings', Icon: SettingsIcon }
];

function Sidebar({ labels, page, language, onNavigate, onLanguageChange }) {
  return (
    <nav className="st-sidebar" aria-label="Sortly">
      <div className="st-sidebar__brand">
        <img src={logo} alt="" />
        <span>Sortly</span>
      </div>

      <div className="st-sidebar__nav">
        {PAGES.map(({ id, labelKey, Icon }) => (
          <button
            key={id}
            type="button"
            className="st-nav"
            aria-current={page === id ? 'page' : undefined}
            onClick={() => onNavigate(id)}
          >
            <Icon />
            {labels[labelKey]}
            <span className="st-nav__mark" />
          </button>
        ))}
      </div>

      <div className="st-sidebar__foot">
        <span className="st-path__label">{labels.language}</span>
        <LanguageToggle language={language} label={labels.language} onChange={onLanguageChange} />
      </div>
    </nav>
  );
}

export default Sidebar;

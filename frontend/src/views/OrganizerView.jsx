import { useState } from 'react';
import NotificationsCenter from '../components/NotificationsCenter';
import PageHead from '../components/PageHead';
import Sidebar from '../components/Sidebar';
import { getCopy } from '../i18n/language';
import organizerCopy from '../i18n/organizerCopy';
import OrganizePage from './OrganizePage';
import SettingsPage from './SettingsPage';

// Página aberta e painel de notificações são estado de UI; o resto vem do
// backend pronto (ADR 0005, #74).
const HEADINGS = {
  organize: ['titleOrganize', 'subtitleOrganize'],
  settings: ['titleSettings', 'subtitleSettings']
};

function OrganizerView({
  language,
  criteria,
  sourceFolderPath,
  destinationFolderPath,
  hasUndo,
  isLoading,
  loadingAction,
  notifications,
  onClearNotifications,
  onLanguageChange,
  onCriterionChange,
  onSelectSourceFolder,
  onSelectDestinationFolder,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
  const [page, setPage] = useState('organize');
  const [isNotificationsOpen, setIsNotificationsOpen] = useState(false);

  const text = getCopy(organizerCopy, language);
  const [titleKey, subtitleKey] = HEADINGS[page];

  const navigate = (next) => {
    setPage(next);
    setIsNotificationsOpen(false);
  };

  return (
    <div className="st-app">
      <Sidebar
        labels={text}
        page={page}
        language={language}
        onNavigate={navigate}
        onLanguageChange={onLanguageChange}
      />

      <main className="st-main">
        <div className="st-main__inner">
          <PageHead title={text[titleKey]} subtitle={text[subtitleKey]}>
            <NotificationsCenter
              labels={text}
              notifications={notifications}
              isOpen={isNotificationsOpen}
              onToggle={() => setIsNotificationsOpen((open) => !open)}
              onClear={onClearNotifications}
            />
          </PageHead>

          {page === 'organize' && (
            <OrganizePage
              labels={text}
              sourceFolderPath={sourceFolderPath}
              destinationFolderPath={destinationFolderPath}
              hasUndo={hasUndo}
              isLoading={isLoading}
              loadingAction={loadingAction}
              onSelectSourceFolder={onSelectSourceFolder}
              onSelectDestinationFolder={onSelectDestinationFolder}
              onOrganizeFiles={onOrganizeFiles}
              onUndoLastOrganization={onUndoLastOrganization}
            />
          )}

          {page === 'settings' && (
            <SettingsPage labels={text} criteria={criteria} onCriterionChange={onCriterionChange} />
          )}
        </div>
      </main>
    </div>
  );
}

export default OrganizerView;

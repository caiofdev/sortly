import { useState } from 'react';
import NotificationsCenter from '../components/NotificationsCenter';
import PageHead from '../components/PageHead';
import Sidebar from '../components/Sidebar';
import ToastStack from '../components/ToastStack';
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
  theme,
  criteria,
  sourceFolderPath,
  destinationFolderPath,
  preview,
  progress,
  lastResult,
  hasUndo,
  isLoading,
  loadingAction,
  notifications,
  unread,
  onClearNotifications,
  onMarkNotificationsRead,
  onLanguageChange,
  onThemeChange,
  onCriterionChange,
  onSelectSourceFolder,
  onSelectDestinationFolder,
  onOrganizeFiles,
  onUndoLastOrganization,
  onCancel,
  onOpenDestination,
  onStartOver
}) {
  const [page, setPage] = useState('organize');
  const [isNotificationsOpen, setIsNotificationsOpen] = useState(false);

  const text = getCopy(organizerCopy, language);
  const [titleKey, subtitleKey] = HEADINGS[page];

  const navigate = (next) => {
    setPage(next);
    setIsNotificationsOpen(false);
  };

  // Abrir o painel já conta como ler, como no protótipo (#75).
  const toggleNotifications = () => {
    if (!isNotificationsOpen) onMarkNotificationsRead();
    setIsNotificationsOpen(!isNotificationsOpen);
  };

  const markRead = () => {
    onMarkNotificationsRead();
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
              unread={unread}
              isOpen={isNotificationsOpen}
              onToggle={toggleNotifications}
              onClose={() => setIsNotificationsOpen(false)}
              onMarkRead={markRead}
              onClear={onClearNotifications}
            />
          </PageHead>

          {page === 'organize' && (
            <OrganizePage
              labels={text}
              sourceFolderPath={sourceFolderPath}
              destinationFolderPath={destinationFolderPath}
              preview={preview}
              progress={progress}
              lastResult={lastResult}
              hasUndo={hasUndo}
              isLoading={isLoading}
              loadingAction={loadingAction}
              onSelectSourceFolder={onSelectSourceFolder}
              onSelectDestinationFolder={onSelectDestinationFolder}
              onOrganizeFiles={onOrganizeFiles}
              onUndoLastOrganization={onUndoLastOrganization}
              onCancel={onCancel}
              onOpenDestination={onOpenDestination}
              onStartOver={onStartOver}
            />
          )}

          {page === 'settings' && (
            <SettingsPage
              labels={text}
              theme={theme}
              criteria={criteria}
              onThemeChange={onThemeChange}
              onCriterionChange={onCriterionChange}
            />
          )}
        </div>
      </main>

      <ToastStack notifications={notifications} closeLabel={text.toastClose} />
    </div>
  );
}

export default OrganizerView;

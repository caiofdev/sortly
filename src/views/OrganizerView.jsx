import { useEffect, useState } from 'react';
import DragDropPanel from '../components/DragDropPanel';
import FolderPathsPanel from '../components/FolderPathsPanel';
import NotificationsCenter from '../components/NotificationsCenter';
import OrganizerActions from '../components/OrganizerActions';
import OrganizerHeader from '../components/OrganizerHeader';
import organizerCopy from '../i18n/organizerCopy';

function OrganizerView({
  language,
  sourceFolderPath,
  destinationFolderPath,
  hasUndo,
  isLoading,
  loadingAction,
  feedback,
  onLanguageChange,
  onResolveDroppedPath,
  onSelectSourceFolder,
  onSelectDestinationFolder,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
  const [notifications, setNotifications] = useState([]);
  const [isNotificationsOpen, setIsNotificationsOpen] = useState(false);

  useEffect(() => {
    if (!feedback) {
      return;
    }

    const now = new Date();
    const formattedTime = now.toLocaleTimeString(language === 'pt-BR' ? 'pt-BR' : 'en-US', {
      hour: '2-digit',
      minute: '2-digit'
    });

    const notificationItem = {
      id: `${Date.now()}-${Math.random().toString(16).slice(2, 8)}`,
      type: feedback.type,
      message: feedback.message,
      time: formattedTime
    };

    setNotifications((previous) => [notificationItem, ...previous].slice(0, 80));
  }, [feedback, language]);

  const text = organizerCopy[language] || organizerCopy['pt-BR'];
  const hasDestination = Boolean(destinationFolderPath);
  const destinationLabel = destinationFolderPath || text.destinationEmpty;

  return (
    <main className="relative min-h-screen overflow-hidden bg-[#0F172A] text-[#F8FAFC]">
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_10%_15%,rgba(59,130,246,0.16),transparent_35%),radial-gradient(circle_at_90%_85%,rgba(34,197,94,0.12),transparent_30%)]" />

      <section className="relative mx-auto flex min-h-screen w-full max-w-4xl items-center justify-center px-6 py-12">
        <div className="w-full rounded-2xl border border-white/10 bg-[#1E293B]/90 p-8 shadow-[0_22px_60px_rgba(2,6,23,0.45)] backdrop-blur-sm md:p-10">
          <OrganizerHeader language={language} onLanguageChange={onLanguageChange} subtitle={text.subtitle} />

          <div className="mt-8 grid gap-5">
            <DragDropPanel
              isLoading={isLoading}
              labels={text}
              onResolveDroppedPath={onResolveDroppedPath}
              onSelectSourceFolder={onSelectSourceFolder}
            />

            <FolderPathsPanel
              sourceLabel={text.sourceLabel}
              destinationLabel={text.destinationLabel}
              sourcePath={sourceFolderPath || text.sourceEmpty}
              destinationPath={destinationLabel}
              hasDestination={hasDestination}
              destinationSelectHintPrefix={text.destinationSelectHintPrefix}
              destinationSelectHintAction={text.destinationSelectHintAction}
              onSelectDestinationFolder={onSelectDestinationFolder}
            />

            <OrganizerActions
              labels={text}
              isLoading={isLoading}
              loadingAction={loadingAction}
              hasUndo={hasUndo}
              hasSource={Boolean(sourceFolderPath)}
              onOrganizeFiles={onOrganizeFiles}
              onUndoLastOrganization={onUndoLastOrganization}
            />
          </div>
        </div>
      </section>

      <NotificationsCenter
        labels={text}
        notifications={notifications}
        isOpen={isNotificationsOpen}
        onToggle={() => setIsNotificationsOpen((open) => !open)}
        onClose={() => setIsNotificationsOpen(false)}
        onClear={() => setNotifications([])}
      />
    </main>
  );
}

export default OrganizerView;

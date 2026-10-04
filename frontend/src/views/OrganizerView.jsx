import { useState } from 'react';
import DragDropPanel from '../components/DragDropPanel';
import FolderPathsPanel from '../components/FolderPathsPanel';
import NotificationsCenter from '../components/NotificationsCenter';
import OrganizerActions from '../components/OrganizerActions';
import OrganizerHeader from '../components/OrganizerHeader';
import { getCopy } from '../i18n/language';
import organizerCopy from '../i18n/organizerCopy';

// Apenas apresentação: recebe estado e ações prontos. O único estado local é
// o de abrir/fechar painéis.
function OrganizerView({
  language,
  organizationOptions,
  sourceFolderPath,
  destinationFolderPath,
  hasUndo,
  isLoading,
  loadingAction,
  notifications,
  onClearNotifications,
  onLanguageChange,
  onOptionChange,
  onSelectSourceFolder,
  onSelectDestinationFolder,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
  const [isNotificationsOpen, setIsNotificationsOpen] = useState(false);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);

  const text = getCopy(organizerCopy, language);
  const destinationLabel = destinationFolderPath || text.destinationEmpty;

  return (
    <main className="relative min-h-screen overflow-hidden bg-[#0B1220] text-[#F8FAFC]">
      {/* Camadas de fundo: gradientes suaves + grid sutil para profundidade */}
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_10%_10%,rgba(59,130,246,0.18),transparent_40%),radial-gradient(circle_at_90%_90%,rgba(34,197,94,0.14),transparent_35%),radial-gradient(circle_at_50%_50%,rgba(168,85,247,0.06),transparent_60%)]" />
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.03]"
        style={{
          backgroundImage:
            'linear-gradient(to right, #ffffff 1px, transparent 1px), linear-gradient(to bottom, #ffffff 1px, transparent 1px)',
          backgroundSize: '48px 48px'
        }}
      />

      <section className="relative mx-auto flex min-h-screen w-full max-w-4xl items-center justify-center px-6 py-12">
        <div className="w-full animate-[fadeInUp_0.5s_ease-out] rounded-3xl border border-white/10 bg-gradient-to-b from-[#1E293B]/95 to-[#182236]/95 p-8 shadow-[0_30px_80px_-20px_rgba(2,6,23,0.6),0_0_0_1px_rgba(255,255,255,0.03)] backdrop-blur-md md:p-10">
          <OrganizerHeader
            language={language}
            onLanguageChange={onLanguageChange}
            subtitle={text.subtitle}
            labels={text}
            settingsOpen={isSettingsOpen}
            onToggleSettings={() => setIsSettingsOpen((open) => !open)}
            options={organizationOptions}
            onOptionChange={onOptionChange}
          />

          <div className="my-7 h-px w-full bg-gradient-to-r from-transparent via-white/10 to-transparent" />

          <div className="grid gap-5">
            <DragDropPanel
              isLoading={isLoading}
              labels={text}
              onSelectSourceFolder={onSelectSourceFolder}
            />

            <FolderPathsPanel
              sourceLabel={text.sourceLabel}
              destinationLabel={text.destinationLabel}
              sourcePath={sourceFolderPath || text.sourceEmpty}
              destinationPath={destinationLabel}
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
        onClear={onClearNotifications}
      />

      <style>{`
        @keyframes fadeInUp {
          from { opacity: 0; transform: translateY(12px); }
          to { opacity: 1; transform: translateY(0); }
        }
      `}</style>
    </main>
  );
}

export default OrganizerView;

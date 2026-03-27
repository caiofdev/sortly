import DragDropPanel from '../components/DragDropPanel';
import FeedbackPanel from '../components/FeedbackPanel';
import FolderPathsPanel from '../components/FolderPathsPanel';
import OrganizerActions from '../components/OrganizerActions';
import OrganizerHeader from '../components/OrganizerHeader';
import organizerCopy from '../i18n/organizerCopy';

function OrganizerView({
  language,
  sourceFolderPath,
  destinationFolderPath,
  hasUndo,
  isLoading,
  feedback,
  onLanguageChange,
  onResolveDroppedPath,
  onSelectSourceFolder,
  onSelectDestinationFolder,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
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
              hasUndo={hasUndo}
              hasSource={Boolean(sourceFolderPath)}
              onOrganizeFiles={onOrganizeFiles}
              onUndoLastOrganization={onUndoLastOrganization}
            />

            

            <FeedbackPanel feedback={feedback} emptyMessage={text.feedbackEmpty} />
          </div>
        </div>
      </section>
    </main>
  );
}

export default OrganizerView;

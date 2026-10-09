import DragDropPanel from '../components/DragDropPanel';
import OrganizerActions from '../components/OrganizerActions';
import OrganizingPanel from '../components/OrganizingPanel';
import PathField from '../components/PathField';
import PreviewSummary from '../components/PreviewSummary';

function OrganizePage({
  labels,
  sourceFolderPath,
  destinationFolderPath,
  preview,
  progress,
  hasUndo,
  isLoading,
  loadingAction,
  onSelectSourceFolder,
  onSelectDestinationFolder,
  onOrganizeFiles,
  onUndoLastOrganization,
  onCancel
}) {
  // Enquanto organiza, o card mostra só o progresso, como no protótipo (#78).
  if (loadingAction === 'organize') {
    return (
      <section className="st-card">
        <OrganizingPanel
          labels={labels}
          progress={progress}
          sourceFolderPath={sourceFolderPath}
          destinationFolderPath={destinationFolderPath}
          onCancel={onCancel}
        />
      </section>
    );
  }

  return (
    <section className="st-card">
      <DragDropPanel
        isLoading={isLoading}
        labels={labels}
        onSelectSourceFolder={onSelectSourceFolder}
      />

      <div className="st-grid2">
        <PathField
          label={labels.sourceLabel}
          path={sourceFolderPath}
          emptyText={labels.sourceEmpty}
        />
        <PathField
          label={labels.destinationLabel}
          path={destinationFolderPath}
          emptyText={labels.destinationEmpty}
          actionLabel={destinationFolderPath ? labels.destinationChange : labels.destinationSelect}
          onAction={onSelectDestinationFolder}
          disabled={isLoading}
        />
      </div>

      <PreviewSummary labels={labels} preview={preview} />

      <OrganizerActions
        labels={labels}
        isLoading={isLoading}
        loadingAction={loadingAction}
        hasUndo={hasUndo}
        hasSource={Boolean(sourceFolderPath)}
        onOrganizeFiles={onOrganizeFiles}
        onUndoLastOrganization={onUndoLastOrganization}
      />
    </section>
  );
}

export default OrganizePage;

import DragDropPanel from '../components/DragDropPanel';
import OrganizerActions from '../components/OrganizerActions';
import PathField from '../components/PathField';

function OrganizePage({
  labels,
  sourceFolderPath,
  destinationFolderPath,
  hasUndo,
  isLoading,
  loadingAction,
  onSelectSourceFolder,
  onSelectDestinationFolder,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
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

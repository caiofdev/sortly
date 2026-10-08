import { LoaderIcon, PlayIcon, UndoIcon } from './icons';

function OrganizerActions({
  labels,
  isLoading,
  loadingAction,
  hasUndo,
  hasSource,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
  const isOrganizing = loadingAction === 'organize';
  const isRestoring = loadingAction === 'restore';

  return (
    <div className="st-actions">
      <button
        type="button"
        className={`st-btn st-btn--primary st-btn--lg${isOrganizing ? ' st-btn--busy' : ''}`}
        onClick={onOrganizeFiles}
        disabled={isLoading || !hasSource}
        aria-busy={isOrganizing}
      >
        {isOrganizing ? <LoaderIcon /> : <PlayIcon />}
        {isOrganizing ? labels.organizing : labels.organize}
      </button>
      <button
        type="button"
        className={`st-btn st-btn--danger st-btn--lg${isRestoring ? ' st-btn--busy' : ''}`}
        onClick={onUndoLastOrganization}
        disabled={isLoading || !hasUndo}
        aria-busy={isRestoring}
      >
        {isRestoring ? <LoaderIcon /> : <UndoIcon />}
        {isRestoring ? labels.restoring : labels.undo}
      </button>
    </div>
  );
}

export default OrganizerActions;

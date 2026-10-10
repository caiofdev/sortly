import { PlayIcon, UndoIcon } from './icons';

// Organizar e desfazer trocam o card pela tela de progresso; aqui só a trava (#78, #96).
function OrganizerActions({
  labels,
  isLoading,
  hasUndo,
  hasSource,
  onOrganizeFiles,
  onUndoLastOrganization
}) {
  return (
    <div className="st-actions">
      <button
        type="button"
        className="st-btn st-btn--primary st-btn--lg"
        onClick={onOrganizeFiles}
        disabled={isLoading || !hasSource}
      >
        <PlayIcon />
        {labels.organize}
      </button>
      <button
        type="button"
        className="st-btn st-btn--danger st-btn--lg"
        onClick={onUndoLastOrganization}
        disabled={isLoading || !hasUndo}
      >
        <UndoIcon />
        {labels.undo}
      </button>
    </div>
  );
}

export default OrganizerActions;

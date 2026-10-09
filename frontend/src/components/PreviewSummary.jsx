// A prévia vem pronta do backend; aqui só se escolhem os textos. A região
// anunciável fica sempre montada, para o leitor de tela ler a contagem quando
// ela chega (#77).
function PreviewContent({ labels, preview }) {
  if (preview.status === 'loading') {
    return <span className="st-preview__label">{labels.previewCounting}</span>;
  }
  if (preview.status !== 'ready') return null;

  return (
    <>
      <span className="st-preview__label">
        <span className="st-preview__count">{preview.totalFiles}</span>{' '}
        {preview.totalFiles === 1 ? labels.previewFoundOne : labels.previewFound}
      </span>
      {preview.folders.map((folder) => (
        <span key={folder.name} className="st-chip">
          {folder.name || labels.previewRoot} <b>{folder.count}</b>
        </span>
      ))}
      {preview.otherFiles > 0 && (
        <span className="st-chip">
          {labels.previewOthers} <b>{preview.otherFiles}</b>
        </span>
      )}
    </>
  );
}

function PreviewSummary({ labels, preview }) {
  return (
    <div className="st-preview" aria-live="polite">
      <PreviewContent labels={labels} preview={preview} />
    </div>
  );
}

export default PreviewSummary;

// Partes opcionais da mensagem: só aparecem quando o valor é maior que zero,
// para a mensagem continuar igual à da versão anterior no caso comum.
const optional = (value, text) => (value > 0 ? ` ${text}: ${value}.` : '');

const feedbackCopy = {
  'pt-BR': {
    sourceSelectError: 'Não foi possível selecionar a pasta de origem.',
    destinationSelectError: 'Não foi possível selecionar a pasta de destino.',
    sourceRequired: 'Selecione a pasta de origem antes de organizar.',
    organizeUnexpectedError: 'Erro inesperado ao organizar os arquivos.',
    undoUnexpectedError: 'Erro inesperado ao desfazer a organização.',
    droppedPathUnexpectedError: 'Não foi possível usar o item arrastado.',
    droppedPathSuccess: 'Origem definida por arrastar e soltar.',
    recoveredLastOrganization: 'Última organização recuperada. Você pode desfazer essa alteração.',
    organizeSuccess: (result) =>
      `Organização concluida: ${result.movedFiles} arquivo(s) movido(s). Origem: ${result.sourceFolderPath}. Destino: ${result.destinationFolderPath}. Processados: ${result.processedFiles}. Ignorados sem extensão: ${result.ignoredWithoutExtension}. Pastas ignoradas: ${result.ignoredFolders}.` +
      optional(result.unchangedFiles, 'Já estavam no lugar') +
      optional(result.failedFiles, 'Falhas ao mover'),
    undoSuccess: (result) =>
      `Desfazer concluido: ${result.restoredFiles} arquivo(s) restaurado(s). Renomeados na restauração: ${result.renamedOnRestore}. Não encontrados: ${result.skippedMissing}.` +
      optional(result.failedFiles, 'Falhas ao restaurar'),
    // Textos dos códigos de erro do backend (ADR 0004). Em português, são os
    // mesmos da versão 1.0.
    errors: {
      INVALID_SOURCE: 'Pasta inválida.',
      INVALID_DESTINATION: 'Pasta de destino inválida.',
      NO_CRITERIA: 'Selecione ao menos um criterio de organizacao.',
      NOTHING_TO_UNDO: 'Nenhuma separação recente para desfazer.',
      DROPPED_INVALID: 'Item arrastado inválido.',
      DROPPED_MISSING: 'O item arrastado não existe mais.',
      DROPPED_UNSUPPORTED: 'Somente arquivos e pastas podem ser arrastados.',
      RECORD_NOT_SAVED:
        'Os arquivos foram organizados, mas não foi possível salvar o registro para desfazer.'
    }
  },
  en: {
    sourceSelectError: 'Could not select the source folder.',
    destinationSelectError: 'Could not select the destination folder.',
    sourceRequired: 'Select a source folder before organizing.',
    organizeUnexpectedError: 'Unexpected error while organizing files.',
    undoUnexpectedError: 'Unexpected error while undoing organization.',
    droppedPathUnexpectedError: 'Could not use the dropped item.',
    droppedPathSuccess: 'Source folder set from drag and drop.',
    recoveredLastOrganization: 'Last organization recovered. You can undo this change.',
    organizeSuccess: (result) =>
      `Organization complete: ${result.movedFiles} file(s) moved. Source: ${result.sourceFolderPath}. Destination: ${result.destinationFolderPath}. Processed: ${result.processedFiles}. Ignored without extension: ${result.ignoredWithoutExtension}. Ignored folders: ${result.ignoredFolders}.` +
      optional(result.unchangedFiles, 'Already in place') +
      optional(result.failedFiles, 'Failed to move'),
    undoSuccess: (result) =>
      `Undo complete: ${result.restoredFiles} file(s) restored. Renamed on restore: ${result.renamedOnRestore}. Missing: ${result.skippedMissing}.` +
      optional(result.failedFiles, 'Failed to restore'),
    errors: {
      INVALID_SOURCE: 'Invalid folder.',
      INVALID_DESTINATION: 'Invalid destination folder.',
      NO_CRITERIA: 'Select at least one organization criterion.',
      NOTHING_TO_UNDO: 'No recent organization to undo.',
      DROPPED_INVALID: 'Invalid dropped item.',
      DROPPED_MISSING: 'Dropped item no longer exists.',
      DROPPED_UNSUPPORTED: 'Only files and folders are supported in drag and drop.',
      RECORD_NOT_SAVED: 'Files were organized, but the undo record could not be saved.'
    }
  }
};

export default feedbackCopy;

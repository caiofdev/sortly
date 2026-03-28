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
      `Organização concluida: ${result.movedFiles} arquivo(s) movido(s). Origem: ${result.sourceFolderPath}. Destino: ${result.destinationFolderPath}. Processados: ${result.processedFiles}. Ignorados sem extensão: ${result.ignoredWithoutExtension}. Pastas ignoradas: ${result.ignoredFolders}.`,
    undoSuccess: (result) =>
      `Desfazer concluido: ${result.restoredFiles} arquivo(s) restaurado(s). Renomeados na restauração: ${result.renamedOnRestore}. Não encontrados: ${result.skippedMissing}.`
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
      `Organization complete: ${result.movedFiles} file(s) moved. Source: ${result.sourceFolderPath}. Destination: ${result.destinationFolderPath}. Processed: ${result.processedFiles}. Ignored without extension: ${result.ignoredWithoutExtension}. Ignored folders: ${result.ignoredFolders}.`,
    undoSuccess: (result) =>
      `Undo complete: ${result.restoredFiles} file(s) restored. Renamed on restore: ${result.renamedOnRestore}. Missing: ${result.skippedMissing}.`
  }
};

export default feedbackCopy;

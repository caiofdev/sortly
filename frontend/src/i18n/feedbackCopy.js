// Avisos curtos, no tom do protótipo: o título vai no painel e no toast; o texto
// traz os detalhes que só aparecem quando o valor é maior que zero (#75).
const when = (value, text) => (value > 0 ? text : '');
const sentences = (...parts) => parts.filter(Boolean).join(' ');
const plural = (n, one, many) => `${n} ${n === 1 ? one : many}`;

const organizeTitlePt = (r) => {
  const organized = plural(r.movedFiles, 'arquivo organizado', 'arquivos organizados');
  if (r.failedFiles > 0) return `${organized}, ${r.failedFiles} com falha`;
  if (r.movedFiles === 0) return 'Nenhum arquivo para organizar';
  return `Pronto! ${organized}`;
};

const organizeTitleEn = (r) => {
  const organized = plural(r.movedFiles, 'file organized', 'files organized');
  if (r.failedFiles > 0) return `${organized}, ${r.failedFiles} failed`;
  if (r.movedFiles === 0) return 'No files to organize';
  return `Done! ${organized}`;
};

const feedbackCopy = {
  'pt-BR': {
    sourceSelectError: 'Não foi possível selecionar a pasta de origem.',
    destinationSelectError: 'Não foi possível selecionar a pasta de destino.',
    sourceRequired: 'Selecione a pasta de origem antes de organizar.',
    organizeUnexpectedError: 'Erro inesperado ao organizar os arquivos.',
    undoUnexpectedError: 'Erro inesperado ao desfazer a organização.',
    droppedPathUnexpectedError: 'Não foi possível usar o item arrastado.',
    settingsSaveError: 'Não foi possível salvar a preferência.',
    unexpectedError: 'Erro inesperado.',
    recovered: {
      title: 'Última organização recuperada',
      text: 'Você pode desfazer essa alteração.'
    },
    sourceDropped: (path) => ({ title: 'Origem definida', text: path }),
    organizeDone: (r) => ({
      title: organizeTitlePt(r),
      text: sentences(
        when(r.movedFiles, `Na pasta ${r.destinationFolderPath}.`),
        when(r.unchangedFiles, `${plural(r.unchangedFiles, 'já estava', 'já estavam')} no lugar.`),
        when(
          r.ignoredWithoutExtension,
          `${plural(r.ignoredWithoutExtension, 'sem extensão ficou', 'sem extensão ficaram')} na origem.`
        ),
        when(
          r.failedFiles,
          `${plural(r.failedFiles, 'não pôde ser movido', 'não puderam ser movidos')}.`
        )
      )
    }),
    undoDone: (r) => ({
      title:
        r.failedFiles > 0
          ? `Organização desfeita, ${r.failedFiles} com falha`
          : 'Organização desfeita',
      text: sentences(
        `${plural(r.restoredFiles, 'arquivo voltou', 'arquivos voltaram')} para a origem.`,
        when(
          r.renamedOnRestore,
          `${plural(r.renamedOnRestore, 'foi renomeado', 'foram renomeados')} para não sobrescrever outro.`
        ),
        when(
          r.skippedMissing,
          `${plural(r.skippedMissing, 'não foi encontrado', 'não foram encontrados')}.`
        ),
        when(r.failedFiles, 'Desfaça de novo para tentar restaurar o resto.')
      )
    }),
    errors: {
      INVALID_SOURCE: 'Pasta inválida.',
      INVALID_DESTINATION: 'Pasta de destino inválida.',
      NO_CRITERIA: 'Selecione ao menos um criterio de organizacao.',
      NOTHING_TO_UNDO: 'Nenhuma separação recente para desfazer.',
      DROPPED_INVALID: 'Item arrastado inválido.',
      DROPPED_MISSING: 'O item arrastado não existe mais.',
      DROPPED_UNSUPPORTED: 'Somente arquivos e pastas podem ser arrastados.',
      RECORD_NOT_SAVED:
        'Os arquivos foram organizados, mas não foi possível salvar o registro para desfazer.',
      LAST_CRITERION: 'Mantenha pelo menos um critério marcado.',
      UNKNOWN_CRITERION: 'Critério de organização desconhecido.',
      INVALID_LANGUAGE: 'Idioma não suportado.',
      INVALID_THEME: 'Tema não suportado.',
      SETTINGS_NOT_SAVED: 'Não foi possível salvar a preferência.'
    }
  },
  en: {
    sourceSelectError: 'Could not select the source folder.',
    destinationSelectError: 'Could not select the destination folder.',
    sourceRequired: 'Select a source folder before organizing.',
    organizeUnexpectedError: 'Unexpected error while organizing files.',
    undoUnexpectedError: 'Unexpected error while undoing organization.',
    droppedPathUnexpectedError: 'Could not use the dropped item.',
    settingsSaveError: 'Could not save the preference.',
    unexpectedError: 'Unexpected error.',
    recovered: {
      title: 'Last organization recovered',
      text: 'You can undo this change.'
    },
    sourceDropped: (path) => ({ title: 'Source set', text: path }),
    organizeDone: (r) => ({
      title: organizeTitleEn(r),
      text: sentences(
        when(r.movedFiles, `In ${r.destinationFolderPath}.`),
        when(r.unchangedFiles, `${plural(r.unchangedFiles, 'was', 'were')} already in place.`),
        when(
          r.ignoredWithoutExtension,
          `${plural(r.ignoredWithoutExtension, 'file without extension', 'files without extension')} stayed in the source.`
        ),
        when(r.failedFiles, `${plural(r.failedFiles, 'file', 'files')} could not be moved.`)
      )
    }),
    undoDone: (r) => ({
      title: r.failedFiles > 0 ? `Organizing undone, ${r.failedFiles} failed` : 'Organizing undone',
      text: sentences(
        `${plural(r.restoredFiles, 'file is', 'files are')} back in the source folder.`,
        when(
          r.renamedOnRestore,
          `${plural(r.renamedOnRestore, 'was', 'were')} renamed to avoid overwriting another.`
        ),
        when(r.skippedMissing, `${plural(r.skippedMissing, 'was', 'were')} not found.`),
        when(r.failedFiles, 'Undo again to try restoring the rest.')
      )
    }),
    errors: {
      INVALID_SOURCE: 'Invalid folder.',
      INVALID_DESTINATION: 'Invalid destination folder.',
      NO_CRITERIA: 'Select at least one organization criterion.',
      NOTHING_TO_UNDO: 'No recent organization to undo.',
      DROPPED_INVALID: 'Invalid dropped item.',
      DROPPED_MISSING: 'Dropped item no longer exists.',
      DROPPED_UNSUPPORTED: 'Only files and folders are supported in drag and drop.',
      RECORD_NOT_SAVED: 'Files were organized, but the undo record could not be saved.',
      LAST_CRITERION: 'Keep at least one criterion selected.',
      UNKNOWN_CRITERION: 'Unknown organization criterion.',
      INVALID_LANGUAGE: 'Unsupported language.',
      INVALID_THEME: 'Unsupported theme.',
      SETTINGS_NOT_SAVED: 'Could not save the preference.'
    }
  }
};

export default feedbackCopy;

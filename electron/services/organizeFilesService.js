const path = require('path');
const { getUniqueDestination } = require('../models/pathModel');

const IMAGE_EXTENSIONS = new Set(['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp']);
const PAGED_EXTENSIONS = new Set(['pdf', 'doc', 'docx', 'odt']);

function normalizeOptions(rawOptions) {
  return {
    byDuration: Boolean(rawOptions?.byDuration),
    byPages: Boolean(rawOptions?.byPages),
    byResolution: Boolean(rawOptions?.byResolution),
    byDate: Boolean(rawOptions?.byDate),
    bySize: Boolean(rawOptions?.bySize),
    byExtension: rawOptions?.byExtension !== false
  };
}

function formatDateFolder(dateValue) {
  const d = new Date(dateValue);
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function formatSizeFolder(sizeInBytes) {
  const sizeMb = Math.max(1, Math.ceil(sizeInBytes / (1024 * 1024)));
  return `${sizeMb}mb`;
}

function buildSegments(options, extension, stats) {
  const segments = [];

  if (options.byExtension) {
    if (!extension) {
      return null;
    }
    segments.push(extension);
  }

  if (options.byDate) {
    segments.push(`date-${formatDateFolder(stats.mtime)}`);
  }

  if (options.bySize) {
    segments.push(`size-${formatSizeFolder(stats.size)}`);
  }

  if (options.byResolution && IMAGE_EXTENSIONS.has(extension)) {
    segments.push('resolution-unknown');
  }

  if (options.byDuration && extension === 'mp4') {
    segments.push('duration-unknown');
  }

  if (options.byPages && PAGED_EXTENSIONS.has(extension)) {
    segments.push('pages-unknown');
  }

  return segments;
}

async function organizeFiles(payload, dependencies) {
  const { fsAdapter, lastOperationRepository } = dependencies;

  const sourceFolderPath = payload?.sourceFolderPath;
  const destinationFolderPath = payload?.destinationFolderPath || sourceFolderPath;
  const organizationOptions = normalizeOptions(payload?.organizationOptions);

  if (!sourceFolderPath || typeof sourceFolderPath !== 'string') {
    throw new Error('Pasta inválida.');
  }

  if (!destinationFolderPath || typeof destinationFolderPath !== 'string') {
    throw new Error('Pasta de destino inválida.');
  }

  const hasAtLeastOneCriterion = Object.values(organizationOptions).some(Boolean);
  if (!hasAtLeastOneCriterion) {
    throw new Error('Selecione ao menos um criterio de organizacao.');
  }

  const entries = await fsAdapter.readDirectoryWithTypes(sourceFolderPath);
  let processedFiles = 0;
  let movedFiles = 0;
  let ignoredWithoutExtension = 0;
  let ignoredFolders = 0;
  const movedItems = [];
  const createdFolders = new Set();

  for (const entry of entries) {
    if (entry.isDirectory()) {
      ignoredFolders += 1;
      continue;
    }

    if (!entry.isFile()) {
      continue;
    }

    processedFiles += 1;

    const sourcePath = path.join(sourceFolderPath, entry.name);
    const extension = path.extname(entry.name).slice(1).toLowerCase();
    const stats = await fsAdapter.getPathStats(sourcePath);

    if (!stats) {
      continue;
    }

    const segments = buildSegments(organizationOptions, extension, stats);
    if (!segments) {
      ignoredWithoutExtension += 1;
      continue;
    }

    const targetFolderPath = path.join(destinationFolderPath, ...segments);
    await fsAdapter.createDirectory(targetFolderPath);
    createdFolders.add(targetFolderPath);

    const destinationPath = path.join(targetFolderPath, entry.name);
    const safeDestinationPath = await getUniqueDestination(destinationPath, fsAdapter.pathExists);

    await fsAdapter.moveFile(sourcePath, safeDestinationPath);
    movedItems.push({ from: sourcePath, to: safeDestinationPath });
    movedFiles += 1;
  }

  lastOperationRepository.setLastOperation({
    sourceFolderPath,
    destinationFolderPath,
    movedItems,
    createdFolders: [...createdFolders]
  });

  return {
    sourceFolderPath,
    destinationFolderPath,
    processedFiles,
    movedFiles,
    ignoredWithoutExtension,
    ignoredFolders,
    message: `Organização concluída: ${movedFiles} arquivo(s) movido(s).`,
    canUndo: movedItems.length > 0
  };
}

module.exports = {
  organizeFiles
};
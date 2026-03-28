const path = require('path');
const fs = require('fs/promises');
const { imageSizeFromFile } = require('image-size/fromFile');
const { PDFDocument } = require('pdf-lib');
const JSZip = require('jszip');
const { getUniqueDestination } = require('../models/pathModel');

const IMAGE_EXTENSIONS = new Set(['png', 'jpg', 'jpeg', 'gif', 'webp', 'bmp']);
const PAGED_EXTENSIONS = new Set(['pdf', 'doc', 'docx', 'odt']);

let musicMetadataModulePromise;

async function getMusicMetadataModule() {
  if (!musicMetadataModulePromise) {
    musicMetadataModulePromise = import('music-metadata');
  }

  return musicMetadataModulePromise;
}

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

async function getResolutionFolderName(filePath) {
  try {
    const dimensions = await imageSizeFromFile(filePath);

    if (!dimensions?.width || !dimensions?.height) {
      return 'unknown';
    }

    return `${dimensions.width}x${dimensions.height}`;
  } catch {
    return 'unknown';
  }
}

function formatDurationFolder(seconds) {
  const safeSeconds = Math.max(0, Math.round(seconds));
  const hours = Math.floor(safeSeconds / 3600);
  const minutes = Math.floor((safeSeconds % 3600) / 60);
  const remainingSeconds = safeSeconds % 60;

  return `duration-${String(hours).padStart(2, '0')}h${String(minutes).padStart(2, '0')}m${String(remainingSeconds).padStart(2, '0')}s`;
}

function parsePositiveInteger(value) {
  const parsed = Number.parseInt(String(value || ''), 10);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : null;
}

async function getDurationFolderName(filePath, extension) {
  if (extension !== 'mp4') {
    return 'duration-unknown';
  }

  try {
    const { parseFile } = await getMusicMetadataModule();
    const metadata = await parseFile(filePath, {
      duration: true,
      skipCovers: true
    });

    if (!Number.isFinite(metadata?.format?.duration) || metadata.format.duration <= 0) {
      return 'duration-unknown';
    }

    return formatDurationFolder(metadata.format.duration);
  } catch {
    return 'duration-unknown';
  }
}

async function getPdfPagesCount(filePath) {
  const buffer = await fs.readFile(filePath);
  const pdfDocument = await PDFDocument.load(buffer, { updateMetadata: false });
  return pdfDocument.getPageCount();
}

async function getDocxPagesCount(filePath) {
  const buffer = await fs.readFile(filePath);
  const zip = await JSZip.loadAsync(buffer);
  const appFile = zip.file('docProps/app.xml');

  if (!appFile) {
    return null;
  }

  const appXml = await appFile.async('string');
  const pagesMatch = appXml.match(/<Pages>(\d+)<\/Pages>/i);
  return parsePositiveInteger(pagesMatch?.[1]);
}

async function getOdtPagesCount(filePath) {
  const buffer = await fs.readFile(filePath);
  const zip = await JSZip.loadAsync(buffer);
  const metaFile = zip.file('meta.xml');

  if (!metaFile) {
    return null;
  }

  const metaXml = await metaFile.async('string');
  const pageCountMatch = metaXml.match(/meta:page-count="(\d+)"/i);
  return parsePositiveInteger(pageCountMatch?.[1]);
}

async function getPagesFolderName(filePath, extension) {
  try {
    let pageCount = null;

    if (extension === 'pdf') {
      pageCount = await getPdfPagesCount(filePath);
    } else if (extension === 'docx') {
      pageCount = await getDocxPagesCount(filePath);
    } else if (extension === 'odt') {
      pageCount = await getOdtPagesCount(filePath);
    }

    if (!Number.isFinite(pageCount) || pageCount <= 0) {
      return 'pages-unknown';
    }

    return `pages-${pageCount}`;
  } catch {
    return 'pages-unknown';
  }
}

async function buildSegments(options, extension, stats, sourcePath) {
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
    segments.push(await getResolutionFolderName(sourcePath));
  }

  if (options.byDuration && extension === 'mp4') {
    segments.push(await getDurationFolderName(sourcePath, extension));
  }

  if (options.byPages && PAGED_EXTENSIONS.has(extension)) {
    segments.push(await getPagesFolderName(sourcePath, extension));
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

    const segments = await buildSegments(
      organizationOptions,
      extension,
      stats,
      sourcePath
    );
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
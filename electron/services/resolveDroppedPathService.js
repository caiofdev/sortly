const path = require('path');

async function resolveDroppedPath(droppedPath, dependencies) {
  const { fsAdapter } = dependencies;

  if (!droppedPath || typeof droppedPath !== 'string') {
    throw new Error('Invalid dropped item.');
  }

  const stats = await fsAdapter.getPathStats(droppedPath);
  if (!stats) {
    throw new Error('Dropped item no longer exists.');
  }

  if (stats.isDirectory()) {
    return { sourceFolderPath: droppedPath };
  }

  if (stats.isFile()) {
    return { sourceFolderPath: path.dirname(droppedPath) };
  }

  throw new Error('Only files and folders are supported in drag and drop.');
}

module.exports = {
  resolveDroppedPath
};

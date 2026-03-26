const path = require('path');

async function getUniqueDestination(destinationPath, pathExists) {
  if (!(await pathExists(destinationPath))) {
    return destinationPath;
  }

  const parentDir = path.dirname(destinationPath);
  const extension = path.extname(destinationPath);
  const baseName = path.basename(destinationPath, extension);
  let counter = 1;

  while (true) {
    const candidatePath = path.join(parentDir, `${baseName} (${counter})${extension}`);
    if (!(await pathExists(candidatePath))) {
      return candidatePath;
    }
    counter += 1;
  }
}

module.exports = {
  getUniqueDestination
};
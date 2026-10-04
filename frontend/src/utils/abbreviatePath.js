export const MAX_PATH_LENGTH = 72;
const KEPT_EDGE_LENGTH = 32;

// Caminhos longos mostram o começo e o fim: "C:\Users\...\Organizados".
export function abbreviatePath(path) {
  if (path.length <= MAX_PATH_LENGTH) {
    return path;
  }
  return `${path.slice(0, KEPT_EDGE_LENGTH)}...${path.slice(-KEPT_EDGE_LENGTH)}`;
}

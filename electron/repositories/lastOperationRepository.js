let lastOperation = null;

function getLastOperation() {
  return lastOperation;
}

function setLastOperation(operation) {
  lastOperation = operation;
}

function clearLastOperation() {
  lastOperation = null;
}

module.exports = {
  getLastOperation,
  setLastOperation,
  clearLastOperation
};
export class LatestRequestGuard {
  private generation = 0;
  private mounted = true;

  begin() {
    const request = ++this.generation;
    return () => this.mounted && request === this.generation;
  }

  invalidate() {
    this.generation += 1;
  }

  unmount() {
    this.mounted = false;
    this.invalidate();
  }
}

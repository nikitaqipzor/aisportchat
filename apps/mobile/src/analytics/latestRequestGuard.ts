export type RequestTicket<Key extends string> = Readonly<{
  key: Key;
  generation: number;
  authEpoch: number;
  mountEpoch: number;
}>;

/**
 * Rejects responses that no longer belong to the current screen, token, or
 * newest request for a section. Keeping it React-free makes the race contract
 * executable in CI instead of relying on timing-sensitive component tests.
 */
export class LatestRequestGuard<Key extends string> {
  private generations = new Map<Key, number>();
  private authEpoch = 0;
  private mountEpoch = 0;
  private mounted = false;

  mount() {
    this.mountEpoch += 1;
    this.mounted = true;
  }

  unmount() {
    this.mountEpoch += 1;
    this.mounted = false;
  }

  rotateAuth() {
    this.authEpoch += 1;
    return this.authEpoch;
  }

  begin(key: Key): RequestTicket<Key> {
    const generation = (this.generations.get(key) ?? 0) + 1;
    this.generations.set(key, generation);
    return {key, generation, authEpoch: this.authEpoch, mountEpoch: this.mountEpoch};
  }

  isLatest(ticket: RequestTicket<Key>) {
    return this.mounted
      && ticket.mountEpoch === this.mountEpoch
      && ticket.authEpoch === this.authEpoch
      && ticket.generation === this.generations.get(ticket.key);
  }
}

export class LatestRefreshOwner {
  private generation = 0;

  begin() { this.generation += 1; return this.generation; }
  release(owner: number) { return owner === this.generation; }
  invalidate() { this.generation += 1; return this.generation; }
}

export type LoadPhase = 'loading' | 'fresh' | 'stale' | 'error' | 'retry';
export type LoadState<T> = {phase: LoadPhase; data: T | null; error: string};

export function beginLoad<T>(previous: LoadState<T>, retry = false): LoadState<T> {
  return {...previous, phase: retry ? 'retry' : previous.data ? 'stale' : 'loading', error: ''};
}

export function failLoad<T>(previous: LoadState<T>, error: string): LoadState<T> {
  return {...previous, phase: previous.data ? 'stale' : 'error', error};
}

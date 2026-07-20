import { useEffect, useRef, useCallback } from 'react';

/**
 * useWailsEvent subscribes to a Wails event and calls the handler.
 * Falls back gracefully if Wails runtime isn't available (dev mode in browser).
 */
export function useWailsEvent(eventName: string, handler: (data: any) => void) {
  const handlerRef = useRef(handler);
  handlerRef.current = handler;

  useEffect(() => {
    const runtime = (window as any).runtime;
    if (!runtime?.EventsOn) return;

    const cancel = runtime.EventsOn(eventName, (...args: any[]) => {
      handlerRef.current(args[0]);
    });

    return () => {
      if (cancel) cancel();
    };
  }, [eventName]);
}

/**
 * useRealtimeProbes listens for probe:result events and appends to a buffer.
 * Returns the latest N results. Much faster than polling.
 */
export function useRealtimeProbes(maxBuffer = 200) {
  const bufferRef = useRef<any[]>([]);
  const listenersRef = useRef<Set<() => void>>(new Set());

  const subscribe = useCallback((listener: () => void) => {
    listenersRef.current.add(listener);
    return () => listenersRef.current.delete(listener);
  }, []);

  useWailsEvent('probe:result', (data) => {
    bufferRef.current = [...bufferRef.current.slice(-(maxBuffer - 1)), data];
    listenersRef.current.forEach((l) => l());
  });

  return { bufferRef, subscribe };
}

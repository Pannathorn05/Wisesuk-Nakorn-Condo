import { useCallback, useEffect, useRef, useState } from "react";

// hook กลางสำหรับ loading/error/empty/success — ทุก section ที่ดึง API ควรใช้ตัวนี้
// แทนการเขียน useState/useEffect ซ้ำเอง (docs/c.md ข้อ 24, .claude/CLAUDE.md)
export function useAsync(fetcher, deps = []) {
  const [state, setState] = useState({ data: null, loading: true, error: null });
  const fetcherRef = useRef(fetcher);
  fetcherRef.current = fetcher;
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    const controller = new AbortController();
    let alive = true;
    setState((s) => ({ ...s, loading: true, error: null }));

    fetcherRef
      .current(controller.signal)
      .then((data) => {
        if (alive) setState({ data, loading: false, error: null });
      })
      .catch((error) => {
        if (error?.name === "AbortError") return;
        if (alive) setState({ data: null, loading: false, error });
      });

    return () => {
      alive = false;
      controller.abort();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [...deps, reloadKey]);

  const refetch = useCallback(() => setReloadKey((k) => k + 1), []);

  return { ...state, refetch };
}

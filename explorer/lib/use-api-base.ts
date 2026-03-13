"use client";

import { useEffect, useState } from "react";

import { API_BASE_STORAGE_KEY, DEFAULT_API_BASE, normalizeApiBase } from "@/lib/api";

export function useApiBase(initialValue?: string) {
  const [apiBase, setApiBase] = useState(() => normalizeApiBase(initialValue ?? DEFAULT_API_BASE));
  const [ready, setReady] = useState(false);

  useEffect(() => {
    const fromQuery = initialValue?.trim();
    const fromStorage = window.localStorage.getItem(API_BASE_STORAGE_KEY);
    const resolved = normalizeApiBase(fromQuery || fromStorage || DEFAULT_API_BASE);
    setApiBase(resolved);
    window.localStorage.setItem(API_BASE_STORAGE_KEY, resolved);
    setReady(true);
  }, [initialValue]);

  function updateApiBase(nextValue: string) {
    const normalized = normalizeApiBase(nextValue);
    setApiBase(normalized);
    window.localStorage.setItem(API_BASE_STORAGE_KEY, normalized);
  }

  return { apiBase, updateApiBase, ready };
}

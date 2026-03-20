"use client";

import { useQuery } from "@tanstack/react-query";
import { apiClient } from "@/lib/api/client";

export function useCurrentUser() {
  return useQuery({
    queryKey: ["session"],
    queryFn: apiClient.me,
    retry: false,
    staleTime: 60_000,
  });
}

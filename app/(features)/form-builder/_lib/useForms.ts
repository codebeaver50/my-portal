"use client";

import { keepPreviousData, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createForm,
  createRecord,
  deleteForm,
  deleteRecord,
  fetchForms,
  fetchRecords,
  updateForm,
} from "./api";
import type { FormInput, RecordInput } from "./types";

const FORMS_KEY = ["form-builder", "forms"];
const recordsKey = (formId: number) => ["form-builder", "records", formId];

export function useForms() {
  return useQuery({
    queryKey: FORMS_KEY,
    queryFn: () => fetchForms(),
  });
}

export function useCreateForm() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: FormInput) => createForm(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: FORMS_KEY });
    },
  });
}

export function useUpdateForm() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: FormInput }) => updateForm(id, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: FORMS_KEY });
    },
  });
}

export function useDeleteForm() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteForm(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: FORMS_KEY });
    },
  });
}

export function useRecords(formId: number, page: number, pageSize: number) {
  return useQuery({
    queryKey: [...recordsKey(formId), page, pageSize],
    queryFn: () => fetchRecords(formId, page, pageSize),
    placeholderData: keepPreviousData,
  });
}

export function useCreateRecord(formId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: RecordInput) => createRecord(formId, input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: recordsKey(formId) });
      queryClient.invalidateQueries({ queryKey: FORMS_KEY });
    },
  });
}

export function useDeleteRecord(formId: number) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (recordId: number) => deleteRecord(formId, recordId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: recordsKey(formId) });
      queryClient.invalidateQueries({ queryKey: FORMS_KEY });
    },
  });
}

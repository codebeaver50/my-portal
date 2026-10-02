export const BOOKS_PAGE_SIZE = 20;

export type BookRow = {
  id: number;
  title: string;
  subtitle: string | null;
  authors: string[];
  publisherName: string | null;
  publishedDate: Date | null;
  coverImageUrl: string | null;
};

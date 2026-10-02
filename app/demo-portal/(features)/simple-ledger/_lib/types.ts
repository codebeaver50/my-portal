export type AccountType = "asset" | "liability" | "equity" | "revenue" | "expense";
export type NormalBalance = "debit" | "credit";
export type EntryType = "debit" | "credit";

export type ChartOfAccount = {
  id: number;
  code: string;
  name: string;
  type: AccountType;
  normalBalance: NormalBalance;
};

export type JournalEntry = {
  id: number;
  chartOfAccountsId: number;
  chartOfAccounts?: ChartOfAccount;
  type: EntryType;
  amount: number;
  description: string;
};

export type Transaction = {
  id: number;
  date: string;
  description: string;
  journalEntries?: JournalEntry[];
  isCorrection: boolean;
  correctedFromId?: number;
  correctionNote: string;
  createdAt: string;
  updatedAt: string;
};

export type JournalEntryInput = {
  chartOfAccountsId: number;
  type: EntryType;
  amount: number;
  description: string;
};

export type TransactionCategoryFilter = "all" | "income" | "expense";

export type TransactionInput = {
  date: string;
  description: string;
  journalEntries: JournalEntryInput[];
  correctionNote?: string;
};

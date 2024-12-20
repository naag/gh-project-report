# Output Format Examples

This document serves as the source of truth for the output format of the project timeline analysis tool. These examples are written considering that two states have been compared.

## No Changes

When no changes are detected:

---

No changes found between 2023-12-01T00:00:00Z and 2024-01-31T00:00:00Z

## Simple Timeline Change

When a task's timeline has changed slightly (less than the threshold for a moderate risk change):

---

Project Timeline Analysis (2023-12-01 → 2024-01-31)

📅 Timeline Changes

| Task   | Status         | Details                         | Start Date | End Date | Duration |
|--------|---------------|--------------------------------|----------------|--------------|--------------|
| Task A | 🔵 On track   | Duration increased by 2 weeks  | Jan 1, 2024    | Feb 14, 2024 | 6 weeks      |
| Task B | 🔵 On track   | Start delayed by 3 weeks, duration decreased by 2 weeks | Jan 22, 2024 | Feb 19, 2024 | 4 weeks |

## Multiple Changes

When multiple tasks have timeline and other changes:

---

Project Timeline Analysis (2023-12-01 → 2024-01-31)

📅 Timeline Changes

| Task   | Status         | Details                         | Start Date | End Date | Duration |
|--------|---------------|--------------------------------|----------------|--------------|--------------|
| Task A | 🔵 On track   | Duration increased by 2 weeks  | Jan 1, 2024    | Feb 14, 2024 | 6 weeks      |
| Task B | 🔵 On track   | Start delayed by 3 weeks, duration decreased by 2 weeks | Jan 22, 2024 | Feb 19, 2024 | 4 weeks |

📋 Other Changes

| Task   | Status           | Priority      | Owner        |
|--------|-----------------|---------------|--------------|
| Task A | 🏗️ In Progress → ✅ Done | - | - |
| Task B | - | High → Medium | Alice → Bob |

## New Item

When a new task is added:

---

Project Timeline Analysis (2023-12-01 → 2024-01-31)

📋 New items

| Task     | Timeline                              | Status     | Priority |
|----------|---------------------------------------|------------|----------|
| New Task | Jan 1, 2024 → Jan 31, 2024 (1 month) | ⏳ Todo    | ⚡ High  |

## High Risk Changes

When tasks have significant delays:

---

Project Timeline Analysis (2023-12-01 → 2024-01-31)

📅 Timeline Changes

| Task   | Status       | Details                         | Start Date | End Date | Duration |
|--------|--------------|--------------------------------|----------------|--------------|--------------|
| Task A | 🔴 High risk | Start delayed by 3 weeks, duration increased by 2 weeks | Jan 22, 2024 | Feb 19, 2024 | 4 weeks |

## Ahead of Schedule

When tasks are ahead of schedule:

---

Project Timeline Analysis (2023-12-01 → 2024-01-31)

📅 Timeline Changes

| Task   | Status              | Details                          | Start Date | End Date | Duration |
|--------|---------------------|--------------------------------|----------------|--------------|--------------|
| Task A | 🚀 Ahead of schedule| Start moved earlier by 2 weeks, duration decreased by 1 week | Jan 22, 2024 | Feb 19, 2024 | 4 weeks |

## Deleted Items

When a task is removed:

---

Project Timeline Analysis (2023-12-01 → 2024-01-31)

📋 Deleted items

| Task         | Priority | Owner | Start Date  | End Date    | Duration |
|--------------|----------|-------|-------------|-------------|----------|
| Removed Task | Medium   | Alice | Jan 1, 2024 | Jan 8, 2024 | 1 week   |


## Format Rules

1. Timeline changes appear before other changes
2. New items only appear in "Other Changes" section
3. Risk levels:
  - 🔵 On track
  - 🟢 Ahead of schedule
  - 🟠 Moderate risk (delay > 14 days)
  - 🔴 High risk (delay > 28 days)
  - 🚫 Added for extreme delays (2x high risk threshold)
4. Dates are formatted as "Jan 2, 2024"
5. Durations use human-readable format:
  - "1 day" for single day
  - "X days" for less than a week
  - "1 week" / "X weeks" for less than 3 months
  - "3 months" / "X months" for longer periods
6. Table columns should be aligned:
  - Left alignment for text columns
  - Right alignment for numeric columns
  - Center alignment for status indicators
7. Redundant information is not included in the output
  - If a date change was already printed, don't print it again
  - If a particular field has no changes, don't print it
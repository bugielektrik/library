# Domain Glossary

> **Business terms and concepts used in this codebase**

## Purpose

This glossary defines business domain terminology so future Claude Code instances (and human developers) understand the MEANING behind the code, not just the implementation.

**When to read:** ALWAYS read this in your first 2 minutes on any task. Understanding domain concepts prevents implementing wrong solutions.

---

## 📚 Core Entities

### Book

A physical or digital item that can be borrowed from the library.

**Key Concepts:**
- A book is NOT the same as a "title" (we can have 3 copies of "The Great Gatsby")
- Books have **status**: available, loaned, reserved, maintenance, lost
- Books can have **metadata**: genre, awards, translations, edition (stored as JSON)
- Books must have valid **ISBN-13** (13 digits with checksum validation)

**Business Rules:**
- A book can only be loaned if status is "available"
- A book can be reserved even if status is "loaned" (for when it's returned)
- A book in "maintenance" status cannot be loaned or reserved

**Example:**
```
Book ID: abc-123
Title: "The Great Gatsby"
Author: F. Scott Fitzgerald
ISBN: 9780743273565
Status: available
Metadata: {"edition": "Scribner", "genre": "Fiction", "awards": ["Modern Library Top 100"]}
```

---

### Member

A person who has registered with the library and can borrow books.

**Key Concepts:**
- Members have **roles**: user, librarian, admin
- Members can have **subscription tiers**: basic, premium, vip
- Members have **borrowing limits** based on subscription tier
- Members can accumulate **late fees** for overdue books

**Subscription Tiers:**
```
Basic:   Max 3 books at once, 14-day loan period
Premium: Max 10 books at once, 30-day loan period, no late fees
VIP:     Unlimited books, 60-day loan period, priority reservations
```

**Business Rules:**
- Members must verify email before borrowing
- Members with outstanding late fees > $10 cannot borrow
- Members can reserve books for 48 hours

**Example:**
```
Member ID: xyz-789
Email: john@example.com
Full Name: John Doe
Role: user
Subscription: premium
Total Late Fees: $3.50
Books Currently Borrowed: 5
```

---

### Author

The writer of one or more books in the library.

**Key Concepts:**
- Authors can have multiple books
- Books can have multiple authors (co-authors)
- Authors have **biography** and **nationality**

**Business Rules:**
- Author name must be unique
- At least one author required per book
- Authors are soft-deleted (marked inactive) to preserve historical data

**Example:**
```
Author ID: author-456
Name: F. Scott Fitzgerald
Biography: "American novelist known for The Great Gatsby..."
Nationality: American
Books: ["The Great Gatsby", "Tender Is the Night", ...]
```

---

### Loan

Represents a book being borrowed by a member.

**Key Concepts:**
- **Loan Date**: When book was borrowed
- **Due Date**: When book must be returned (loan date + tier duration)
- **Return Date**: When book was actually returned (null if not returned)
- **Status**: active, returned, overdue, lost

**Business Rules:**
- Due date = Loan date + member's subscription tier duration
- If today > due date and return date is null, loan is **overdue**
- Late fee = $0.50 per day overdue (unless premium/vip member)
- Max late fee = $25 (then book is marked "lost")
- When book is returned, update book status to "available"

**Example:**
```
Loan ID: loan-111
Book ID: abc-123 ("The Great Gatsby")
Member ID: xyz-789 (John Doe)
Loan Date: 2024-01-01
Due Date: 2024-01-31 (30 days, premium member)
Return Date: null
Status: active
Late Fee: $0.00
```

**Overdue Example:**
```
Loan ID: loan-222
Book ID: abc-456 ("1984")
Member ID: xyz-999 (Jane Smith - basic member)
Loan Date: 2024-01-01
Due Date: 2024-01-15 (14 days, basic member)
Return Date: null
Status: overdue
Days Overdue: 20 (today is 2024-02-04)
Late Fee: $10.00 (20 days × $0.50/day)
```

---

### Subscription

Represents a member's subscription tier and benefits.

**Key Concepts:**
- Subscription **tier**: basic, premium, vip
- **Start date** and **end date**
- **Auto-renewal**: whether subscription renews automatically
- **Status**: active, expired, canceled

**Pricing (monthly):**
```
Basic:   Free
Premium: $9.99/month
VIP:     $24.99/month
```

**Business Rules:**
- All new members start with "basic" (free)
- Subscriptions can be upgraded mid-period (pro-rated)
- When subscription expires, member reverts to "basic"
- Active loans continue even if subscription expires

**Example:**
```
Subscription ID: sub-333
Member ID: xyz-789
Tier: premium
Start Date: 2024-01-01
End Date: 2024-02-01
Status: active
Auto-Renewal: true
```

---

## 💼 Business Processes

### Borrowing Workflow

```
1. Member selects book (status must be "available")
2. System checks:
   - Member has verified email
   - Member is not at borrowing limit
   - Member has no outstanding late fees > $10
   - Book is available
3. System creates Loan record:
   - Loan date = today
   - Due date = today + member tier duration
   - Status = active
4. System updates Book status to "loaned"
5. Member receives confirmation email
```

**Code location:** `internal/usecase/loanops/create_loan.go`

---

### Returning Workflow

```
1. Member returns book (or librarian processes return)
2. System finds active Loan for book
3. System calculates late fee:
   - Days overdue = today - due date (if positive)
   - Late fee = days overdue × $0.50 (unless premium/vip)
   - Max $25
4. System updates Loan:
   - Return date = today
   - Status = returned
   - Late fee calculated
5. System updates Book status to "available"
6. If late fee > 0, add to member's total late fees
```

**Code location:** `internal/usecase/loanops/return_book.go`

---

### Subscription Upgrade Workflow

```
1. Member requests upgrade (basic → premium or premium → vip)
2. System calculates pro-rated cost
3. Member pays (payment processing outside scope)
4. System creates new Subscription record:
   - Start date = today
   - End date = today + 30 days
   - Tier = new tier
5. System updates Member's subscription reference
6. New benefits apply immediately
```

**Code location:** `internal/usecase/subops/subscribe_member.go`

---

## 📏 Business Rules Reference

### Late Fees

| Member Tier | Late Fee | Max Fee | Grace Period |
|-------------|----------|---------|--------------|
| Basic       | $0.50/day | $25 | None |
| Premium     | $0/day | $0 | Unlimited |
| VIP         | $0/day | $0 | Unlimited |

**Special Cases:**
- If late fee reaches $25, book is marked "lost" and removed from loan
- Member must pay replacement cost (book price + $10 processing fee)
- Lost books are removed from member's borrowing count

---

### Borrowing Limits

| Member Tier | Max Concurrent Loans | Loan Duration | Reservation Priority |
|-------------|---------------------|---------------|---------------------|
| Basic       | 3 | 14 days | Normal |
| Premium     | 10 | 30 days | High |
| VIP         | Unlimited | 60 days | Highest |

**Special Cases:**
- Children's books (age < 12 category) have 7-day loan period regardless of tier
- Reference books cannot be loaned (in-library use only)

---

### Reservation Rules

- **Reservation Hold Time:** 48 hours after book becomes available
- **Queue System:** FIFO (first in, first out)
- **VIP Priority:** VIP members skip to front of queue
- **Expiration:** If not picked up in 48 hours, next person in queue is notified
- **Max Reservations:** Basic=1, Premium=5, VIP=Unlimited

---

## 🔧 Technical Terms (Domain-Specific)

### ISBN (International Standard Book Number)

- **Format:** 13 digits
- **Example:** 9780743273565
- **Validation:** Must pass ISBN-13 checksum algorithm
- **Business Rule:** Each book copy has same ISBN, but different book ID

**Code location:** `internal/domain/book/service.go` → `ValidateISBN()`

---

### Checksum Validation

Algorithm for ISBN-13:
```
1. Sum digits alternating ×1 and ×3 (first 12 digits)
2. Check digit = (10 - (sum % 10)) % 10
3. Must match 13th digit
```

**Why it matters:** Prevents typos when entering ISBNs manually.

---

### Soft Delete

Instead of permanently deleting records, we mark them as "inactive" or "deleted".

**Why:**
- Preserve historical data (e.g., which books a member borrowed)
- Allow "undelete" if mistake was made
- Maintain referential integrity

**Implementation:**
```sql
-- Don't do this:
DELETE FROM authors WHERE id = 'author-123';

-- Do this:
UPDATE authors SET deleted_at = NOW() WHERE id = 'author-123';

-- Queries exclude soft-deleted:
SELECT * FROM authors WHERE deleted_at IS NULL;
```

---

### Pro-rated Pricing

When a member upgrades subscription mid-period, they pay only for remaining days.

**Formula:**
```
Days remaining = End date - Today
Daily rate = Monthly price / 30
Pro-rated cost = Days remaining × Daily rate

Example:
Member has premium ($9.99/month)
Upgrades to VIP ($24.99/month) with 15 days left

VIP daily rate = $24.99 / 30 = $0.833/day
Premium daily rate = $9.99 / 30 = $0.333/day
Difference = $0.50/day

Cost = 15 days × $0.50 = $7.50
```

**Code location:** `internal/domain/subscription/service.go` → `CalculateProRatedCost()`

---

## 🎯 Common Confusions

### Book vs. Book Copy

**WRONG:** "A book can only be borrowed once"
**RIGHT:** "A book COPY can only be borrowed once. We can have multiple copies of the same book."

In our system:
- **Book** = A specific physical/digital copy (has unique ID)
- **Title** = The work (e.g., "The Great Gatsby")
- We track books, not titles (though we could add a "Title" entity in future)

---

### Member vs. User

We use **Member** (not User) because:
- Members belong to a library (member of the library)
- "User" is too generic (user of the API? user of the website?)
- Domain-Driven Design principle: use ubiquitous language

**Always use "Member" in:**
- Entity names (`member.Entity`)
- Table names (`members`)
- API endpoints (`/members`)
- Business discussions

---

### Loan vs. Borrow

- **Loan** (noun): The record of a book being borrowed
- **Borrow** (verb): The act of taking a book
- We store "Loans" in the database
- Members "borrow" books (which creates a Loan)

**Examples:**
```
"Member borrowed a book" → Creates a Loan
"Active loans" → Loans with status=active
"Loan period" → Duration between loan_date and due_date
```

---

### Overdue vs. Late

Same meaning in our domain. We use **overdue**.

```
Loan is overdue = today > due_date AND return_date IS NULL
```

---

## 📖 Domain Events (Future)

We don't currently implement domain events, but these would be valuable:

- `BookBorrowed` → Trigger email notification
- `BookReturned` → Update availability, notify reservation queue
- `BookOverdue` → Send reminder email
- `SubscriptionExpired` → Downgrade to basic tier
- `LateFeeExceeded` → Suspend borrowing privileges

**Future ADR:** When we add event-driven architecture, create ADR explaining why.

---

## 🔗 Cross-References

**To understand implementations, see:**
- **Entities:** `internal/domain/{entity}/entity.go`
- **Business Logic:** `internal/domain/{entity}/service.go`
- **Use Cases:** `internal/usecase/{entity}ops/`
- **Database Schema:** `migrations/postgres/*.up.sql`

**To understand WHY we do things this way:**
- **Clean Architecture:** [adrs/001-clean-architecture.md](./adrs/001-clean-architecture.md)
- **Domain Services:** [adrs/002-domain-services.md](./adrs/002-domain-services.md)

---

## 💡 Pro Tips

1. **When adding a feature, ask:** "What's the business rule?"
   - Don't just implement CRUD
   - Understand WHY the rule exists
   - Check with domain experts if unsure

2. **When naming things, use domain language:**
   - ✅ `CalculateLateFee(daysOverdue int) float64`
   - ❌ `CalculateFee(days int) float64`

3. **When writing tests, use domain scenarios:**
   - ✅ `TestBorrowBook_WhenMemberHasOutstandingLateFees_ReturnsError`
   - ❌ `TestCreate_Invalid_ReturnsError`

4. **When in doubt, check migrations:**
   - Database schema reflects business rules (constraints, foreign keys)
   - `migrations/postgres/*.up.sql` shows what's actually enforced

---

**Last Updated:** 2025-01-19

**Next Review:** When adding new domain entities or changing business rules

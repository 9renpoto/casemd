# Human-first Scenario Test

## Account Management
### Profile Editing
#### Save a profile containing detailed contact information

1. Sign in with an account that can edit its profile and open the profile settings page from the account menu
2. Enter a full postal address, a preferred display name, and a biography long enough to wrap across several lines in a spreadsheet cell
3. Save the profile and reload the page to confirm that the stored information remains available
* [ ] The saved display name appears in the page header and account menu
* [ ] The postal address preserves its intended punctuation and spacing
* [ ] The biography remains readable without losing words or line breaks

#### Cancel an unfinished profile change

1. Replace the current display name with a temporary value
2. Select Cancel and return to the profile overview
* [ ] The original display name remains unchanged
* [ ] No success message appears for the cancelled update

## Order Review
### Checkout Summary
#### Review a multi-item order before payment

1. Add several products with different quantities to the cart
2. Open checkout and compare each line item with the cart
3. Return to the cart, change one quantity, and reopen checkout
* [ ] Product names, quantities, and line totals match the latest cart contents
* [ ] Shipping, tax, discounts, and the final total are clearly separated
* [ ] The payment action remains available after the order summary updates

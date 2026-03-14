# Privacy Policy

**Effective Date:** March 14, 2026
**Last Updated:** March 14, 2026
**Company:** ShadowScale (shadowscale.dev)
**Product:** Relix (relix.sh)

---

## The Short Version

Relix is built on a simple principle: **your code is yours, and we can't read it.** All session data between your AI coding agents and the Relix mobile app is end-to-end encrypted. Our relay server only sees ciphertext — we cannot decrypt your code, your AI conversations, or your tool use details, even if compelled by law.

This privacy policy explains what data we *do* collect, why, and what rights you have.

---

## 1. What We Collect

### Account Information

When you create a Relix account, we collect:

- **Email address** — for authentication, billing receipts, and account recovery
- **Display name** — shown in the app and to team members (if applicable)
- **GitHub username** — if you sign up via GitHub OAuth

### Machine Information

When you pair a machine with Relix, we collect:

- **Machine name** — the hostname you assign (e.g., "Work Laptop")
- **Operating system and architecture** — to ensure compatibility (e.g., "macOS arm64")
- **Agent version** — the version of relixctl installed
- **Public encryption key** — used for end-to-end encryption key exchange

### Session Metadata

For each AI coding session routed through Relix, we collect:

- **Session start and end timestamps**
- **Session status** (active, idle, waiting for approval, completed)
- **Machine identifier** — which paired machine the session is on
- **Message counts** — number of messages routed (not their content)

### Device Information (Mobile App)

- **Device token** — for Apple Push Notification service (APNs) or Firebase Cloud Messaging (FCM)
- **Device type and OS version** — for push notification delivery and compatibility
- **App version** — for support and compatibility

### Billing Information

- **Subscription tier and status** — managed through Stripe
- **Payment method (last 4 digits only)** — we do not store full card numbers; Stripe handles all payment processing

### Usage Analytics

- **Feature usage events** — such as "approval sent," "session viewed," "machine paired"
- **Error reports** — crash logs and error events to improve reliability

### What We Do NOT Collect

- **Your code** — session content is end-to-end encrypted; we cannot read it
- **AI conversations** — all messages between you and your AI agents are encrypted
- **Tool use details** — file names, diffs, and command outputs are encrypted
- **Terminal output** — everything displayed in your session is encrypted
- **Browsing history or contacts** — we don't access anything outside the Relix app

---

## 2. How We Use Your Data

| Data | Purpose |
|------|---------|
| Email address | Authentication, billing receipts, account recovery, service updates |
| Machine name and metadata | Display in your dashboard, troubleshooting |
| Session metadata | Enforcing tier limits (session counts), displaying session history |
| Device tokens | Delivering push notifications when your AI agent needs attention |
| Billing information | Processing subscription payments through Stripe |
| Usage analytics | Improving the product, identifying and fixing bugs |

We do not sell your data. We do not use your data for advertising. We do not share your data with third parties for their marketing purposes.

---

## 3. Third-Party Services

We use the following third-party services to operate Relix:

| Service | Purpose | Their Privacy Policy |
|---------|---------|---------------------|
| **Stripe** | Payment processing and subscription management | [stripe.com/privacy](https://stripe.com/privacy) |
| **Apple Push Notification service (APNs)** | Push notifications to iOS devices | [apple.com/privacy](https://www.apple.com/privacy/) |
| **Firebase Cloud Messaging (FCM)** | Push notifications to Android devices | [firebase.google.com/support/privacy](https://firebase.google.com/support/privacy) |
| **GitHub** | OAuth authentication (optional) | [docs.github.com/en/site-policy/privacy-policies](https://docs.github.com/en/site-policy/privacy-policies) |
| **Fly.io** | Infrastructure hosting | [fly.io/legal/privacy-policy](https://fly.io/legal/privacy-policy/) |

These services receive only the minimum data necessary to perform their function. For example, Stripe receives your email and payment details but never your session data.

---

## 4. Data Retention

| Data | Retention Period |
|------|-----------------|
| Account information | Until you delete your account |
| Machine records | Until you unpair the machine or delete your account |
| Session metadata | 7 days (Free), 30 days (Plus), 90 days (Pro/Team) — then permanently deleted |
| Encrypted session content | Same as session metadata — and we cannot read it regardless |
| Billing records | As required by tax law (typically 7 years for transaction records) |
| Usage analytics | 12 months, then aggregated and anonymized |

When you delete your account, we delete all your personal data within 30 days, except billing records required by law.

---

## 5. Data Security

- All session data is **end-to-end encrypted** using NaCl box (X25519 + XSalsa20-Poly1305). The relay server handles only ciphertext.
- Account data is encrypted at rest and in transit (TLS 1.3).
- Passwords are hashed with bcrypt.
- API authentication uses short-lived JWTs with refresh token rotation.
- The relay and agent source code are open source (MIT license) — you can audit every byte that touches your machine.

---

## 6. Your Rights

### For All Users

You have the right to:

- **Access** your data — export your account information and session metadata at any time from Settings > Privacy > Export Data
- **Correct** your data — update your email, display name, or machine names in the app
- **Delete** your data — delete your account from Settings > Account > Delete Account, or email support@shadowscale.dev
- **Object** to processing — opt out of non-essential analytics in Settings > Privacy

### GDPR Rights (European Economic Area)

If you are in the EEA, you additionally have the right to:

- **Rectification** — correct inaccurate personal data
- **Erasure** ("right to be forgotten") — request deletion of your personal data
- **Restriction** — request that we limit processing of your data
- **Portability** — receive your data in a structured, machine-readable format
- **Withdraw consent** — where processing is based on consent, withdraw at any time
- **Lodge a complaint** — with your local data protection authority

Our legal basis for processing is:
- **Contract performance** — account data, session routing, billing
- **Legitimate interest** — usage analytics, security monitoring, product improvement
- **Consent** — marketing emails (opt-in only)

To exercise any GDPR right, email support@shadowscale.dev. We will respond within 30 days.

### CCPA Rights (California)

If you are a California resident, you have the right to:

- **Know** what personal information we collect and how we use it
- **Delete** your personal information
- **Opt out** of the sale of your personal information — we do not sell personal information
- **Non-discrimination** — we will not treat you differently for exercising your rights

To exercise any CCPA right, email support@shadowscale.dev.

---

## 7. Children

Relix is not intended for use by anyone under the age of 13. We do not knowingly collect personal information from children under 13. If we learn that we have collected data from a child under 13, we will delete it promptly. If you believe a child under 13 has provided us with personal information, please contact us at support@shadowscale.dev.

---

## 8. Changes to This Policy

We may update this privacy policy from time to time. When we make material changes, we will:

- Update the "Last Updated" date at the top
- Notify you by email or in-app notification at least 14 days before the changes take effect
- Post the updated policy at relix.sh/privacy

---

## 9. Contact Us

For privacy questions, data requests, or concerns:

- **Email:** support@shadowscale.dev
- **Website:** relix.sh
- **GitHub:** github.com/relixdev

We aim to respond to all privacy inquiries within 7 business days.

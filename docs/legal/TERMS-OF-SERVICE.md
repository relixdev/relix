# Terms of Service

**Effective Date:** March 14, 2026
**Last Updated:** March 14, 2026
**Company:** ShadowScale (shadowscale.dev)
**Product:** Relix (relix.sh)

---

## 1. What This Agreement Covers

These Terms of Service ("Terms") govern your use of the Relix platform, including the mobile application, the relixctl command-line agent, the relay service, and the cloud API (collectively, the "Service"). By creating an account or using any part of the Service, you agree to these Terms.

Relix is operated by ShadowScale ("we," "us," "our"). If you have questions, contact us at legal@shadowscale.dev.

---

## 2. The Service

Relix is a mobile command center for AI coding agents. It connects AI coding tools (such as Claude Code, Aider, and Cline) running on your machines to a mobile app, enabling:

- Real-time session monitoring across multiple machines
- Push notifications when an AI agent needs your input
- One-tap approval or denial of tool use requests
- Sending messages to active AI coding sessions from your phone
- End-to-end encrypted communication between your machines and mobile device

The relay and command-line agent (relixctl) are open source under the MIT license. The mobile app and cloud API are proprietary.

---

## 3. Your Account

### Creating an Account

You must create an account to use Relix. You can sign up with email/password or GitHub OAuth. You must:

- Be at least 13 years old
- Provide accurate information
- Keep your login credentials secure
- Notify us immediately if you suspect unauthorized access (email support@shadowscale.dev)

### Account Responsibility

You are responsible for all activity under your account, including actions taken by AI coding agents that you approve through Relix. When you tap "Allow" on an approval request, you are authorizing that action on your machine.

---

## 4. Acceptable Use

You agree not to:

- Use Relix to facilitate illegal activity
- Attempt to bypass encryption, authentication, or tier limits
- Reverse-engineer the cloud API or mobile app (the open-source relay and agent are excluded from this restriction)
- Interfere with the relay service or other users' connections
- Use automated systems to create accounts or abuse free tier limits
- Resell access to the Service without our written permission
- Transmit malware or attempt to compromise other users' machines through the relay

We reserve the right to suspend or terminate accounts that violate these terms.

---

## 5. Subscription and Billing

### Plans

Relix offers the following subscription tiers:

| Tier | Price | Includes |
|------|-------|----------|
| **Free** | $0 | 1 machine, 2 concurrent sessions, 7-day history |
| **Plus** | $4.99/month | 3 machines, 5 concurrent sessions, 30-day history |
| **Pro** | $14.99/month | 10 machines, unlimited sessions, 90-day history, priority relay |
| **Team** | $24.99/user/month | Unlimited machines, shared sessions, admin controls, audit log |

Annual plans are available at a discount (2 months free).

### Payment Processing

All payments are processed by Stripe. By subscribing, you also agree to [Stripe's Terms of Service](https://stripe.com/legal). We do not store your full payment card details.

### Auto-Renewal

Subscriptions renew automatically at the end of each billing cycle (monthly or annually). You will be charged the current rate for your plan unless you cancel before the renewal date.

### Cancellation

You can cancel your subscription at any time from the app (Settings > Subscription > Cancel) or by emailing support@shadowscale.dev. When you cancel:

- Your subscription remains active until the end of the current billing period
- You are not charged again after cancellation
- Your account reverts to the Free tier when the paid period ends
- Session history beyond the Free tier retention period (7 days) will be deleted 30 days after downgrade

### Price Changes

We may change subscription prices with at least 30 days' notice. Price changes do not affect your current billing cycle. If you disagree with a price change, you can cancel before the new price takes effect.

---

## 6. Refund Policy

We offer refunds as follows:

- **14-day money-back guarantee:** If you are unsatisfied with a paid subscription, request a full refund within 14 days of your initial purchase or renewal. No questions asked.
- **Pro-rata refunds for annual plans:** If you cancel an annual subscription after 14 days, you may request a pro-rata refund for the unused portion of your term.
- **No refunds** for partial months on monthly plans after the 14-day window.

To request a refund, email support@shadowscale.dev with your account email. Refunds are processed within 5-10 business days to your original payment method.

See our full [Refund Policy](./REFUND-POLICY.md) for details.

---

## 7. Intellectual Property

### Your Content

You own your code, your AI conversations, and all content generated in your coding sessions. Relix does not claim any rights to your content. Because session data is end-to-end encrypted, we cannot access your content even if we wanted to.

### Our Platform

Relix, the mobile app, the cloud API, and the ShadowScale brand are our intellectual property. The relay (relixdev/relay) and CLI agent (relixdev/relixctl) are MIT-licensed open source — you can use, modify, and distribute them under that license.

### Feedback

If you provide feedback or suggestions about Relix, we may use them to improve the Service without obligation to you.

---

## 8. End-to-End Encryption

### What It Means

All session data — messages, tool use requests, approvals, terminal output — is encrypted on your machine before it reaches the relay. The relay routes encrypted blobs. We cannot decrypt your session content.

### What This Means for You

- **We cannot recover your data.** If you lose your encryption keys (e.g., by unpairing all devices without backing up), your encrypted session history is permanently inaccessible. This is by design.
- **We cannot comply with content-based requests.** If law enforcement requests your session content, we can only provide encrypted ciphertext and the metadata described in our Privacy Policy.
- **You are responsible for your approvals.** Since we cannot see what you're approving, you are solely responsible for reviewing tool use requests before tapping Allow.

### Self-Hosting

If you run your own relay, we have even less visibility — we only see your authentication and push notification requests.

---

## 9. Service Availability

### Best-Effort Service

We strive for high availability but do not guarantee uninterrupted service. Specifically:

- **Free tier:** Best-effort availability. No uptime commitment.
- **Plus and Pro tiers:** We target 99.9% uptime for the relay and cloud services, but this is a goal, not a contractual SLA.
- **Team tier:** Includes a 99.9% uptime SLA. Contact sales@shadowscale.dev for details.

### Planned Maintenance

We will provide at least 24 hours' notice for planned maintenance that may affect service availability, via email and in-app notification.

### Service Changes

We may modify, update, or discontinue features of the Service. For material changes that reduce functionality, we will provide at least 30 days' notice. If you disagree with a material change, you can cancel and receive a pro-rata refund for any prepaid period.

---

## 10. Limitation of Liability

### Disclaimer of Warranties

The Service is provided "as is" and "as available." To the maximum extent permitted by law, we disclaim all warranties, express or implied, including warranties of merchantability, fitness for a particular purpose, and non-infringement.

### Limitation

To the maximum extent permitted by law, ShadowScale's total liability to you for any claims arising from or related to the Service is limited to the greater of:

- The amount you paid us in the 12 months before the claim arose, or
- $50

### Exclusions

We are not liable for:

- Actions taken by AI coding agents that you approved through Relix
- Data loss resulting from encryption key loss (see Section 8)
- Third-party service outages (Stripe, APNs, FCM, GitHub)
- Indirect, incidental, consequential, or punitive damages

### Exceptions

Nothing in these Terms limits our liability for fraud, gross negligence, or any liability that cannot be excluded by law.

---

## 11. Indemnification

You agree to indemnify and hold harmless ShadowScale from claims arising from:

- Your violation of these Terms
- Your use of the Service, including actions you approve through AI agents
- Your violation of any applicable law

---

## 12. Termination

### By You

You can delete your account at any time from Settings > Account > Delete Account, or by emailing support@shadowscale.dev. Upon deletion, we delete your personal data within 30 days (except billing records required by law).

### By Us

We may suspend or terminate your account if you:

- Violate these Terms
- Use the Service for illegal purposes
- Abuse the relay or interfere with other users

For non-urgent violations, we will notify you and give you 7 days to resolve the issue before termination. For urgent violations (security threats, illegal activity), we may act immediately.

### Effect of Termination

Upon termination:

- Your access to the Service stops immediately
- Your encrypted session data is deleted within 30 days
- Any active subscription is cancelled (with a pro-rata refund for annual plans if we terminate without cause)
- Account data is handled per our Privacy Policy

---

## 13. Dispute Resolution

### Informal Resolution First

Before filing any legal claim, you agree to contact us at legal@shadowscale.dev and attempt to resolve the dispute informally for at least 30 days.

### Governing Law

These Terms are governed by the laws of the State of California, United States, without regard to conflict of law principles.

### Jurisdiction

Any disputes not resolved informally will be resolved in the state or federal courts located in San Francisco County, California.

---

## 14. General

- **Entire Agreement:** These Terms, together with our Privacy Policy and Refund Policy, constitute the entire agreement between you and ShadowScale regarding the Service.
- **Severability:** If any provision is found unenforceable, the remaining provisions remain in effect.
- **No Waiver:** Our failure to enforce a provision does not waive our right to enforce it later.
- **Assignment:** You may not assign your rights under these Terms. We may assign ours in connection with a merger, acquisition, or sale of assets.
- **Updates:** We may update these Terms with at least 30 days' notice. Continued use after the effective date constitutes acceptance.

---

## 15. Contact

- **General:** support@shadowscale.dev
- **Legal:** legal@shadowscale.dev
- **Website:** relix.sh

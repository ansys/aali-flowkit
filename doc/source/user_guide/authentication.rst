.. _authentication:

Authentication and Access Control
=================================

FlowKit provides comprehensive authentication and access control mechanisms to secure API access, track usage, and manage user permissions across the platform.

Overview
--------

The Authentication system enables:

- API key validation and management
- User creation and identification
- Token usage tracking and limits
- Access denial and notification workflows
- Support for multiple authentication backends (MongoDB, Key-Value stores)
- Email notifications for security events

Architecture
------------

The authentication flow follows a multi-layer approach:

.. code-block:: text

   API Request → Auth Check → User Validation → Token Tracking → Access Decision
        ↓             ↓              ↓                ↓              ↓
   Extract Key   Verify Key    Create/Find    Update Usage    Allow/Deny
                               User ID         Check Limits

Core Functions
--------------

Authentication
~~~~~~~~~~~~~~

.. list-table:: Auth Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``CheckApiKeyAuthMongoDb``
     - Validate API key against MongoDB backend
   * - ``CheckApiKeyAuthKvDb``
     - Validate API key against Key-Value store
   * - ``CheckCreateUserIdMongoDb``
     - Create or retrieve user ID for authenticated requests

Usage Tracking
~~~~~~~~~~~~~~

.. list-table:: Token Tracking Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``UpdateTotalTokenCountForCustomerMongoDb``
     - Track token usage by customer
   * - ``UpdateTotalTokenCountForUserIdMongoDb``
     - Track token usage by user ID
   * - ``UpdateTotalTokenCountForCustomerKvDb``
     - Track usage in Key-Value store

Access Control
~~~~~~~~~~~~~~

.. list-table:: Access Control Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``DenyCustomerAccessAndSendWarningMongoDb``
     - Deny access and notify customer
   * - ``DenyCustomerAccessAndSendWarningMongoDbUserId``
     - Deny access by user ID
   * - ``DenyCustomerAccessAndSendWarningKvDb``
     - Deny access using KV store

Notifications
~~~~~~~~~~~~~

.. list-table:: Notification Functions
   :header-rows: 1
   :widths: 35 65

   * - Function
     - Description
   * - ``SendLogicAppNotificationEmail``
     - Send security notifications via email
   * - ``CreateMessageWithVariable``
     - Create templated notification messages

Configuration
-------------

Configure authentication settings:

.. code-block:: yaml

   # MongoDB Authentication
   MONGO_CONNECTION_STRING: "mongodb://localhost:27017"
   AUTH_DATABASE: "flowkit_auth"
   AUTH_COLLECTION: "api_keys"
   
   # Key-Value Store Authentication
   KV_STORE_TYPE: "redis"
   KV_STORE_ENDPOINT: "redis://localhost:6379"
   
   # Token Limits
   DEFAULT_TOKEN_LIMIT: 1000000
   TOKEN_WINDOW: "monthly"
   
   # Notification Settings
   SMTP_SERVER: "smtp.company.com"
   SECURITY_EMAIL: "security@company.com"
   ENABLE_NOTIFICATIONS: true

Usage Examples
--------------

**Example 1: API Key Validation**

.. code-block:: json

   {
     "name": "CheckApiKeyAuthMongoDb",
     "inputs": {
       "apiKey": "sk-proj-abc123...",
       "requestMetadata": {
         "ip": "192.168.1.100",
         "userAgent": "FlowKit-Client/1.0",
         "timestamp": "2024-01-15T10:30:00Z"
       }
     }
   }

**Example 2: Track Token Usage**

.. code-block:: json

   {
     "name": "UpdateTotalTokenCountForCustomerMongoDb",
     "inputs": {
       "customerId": "cust_12345",
       "tokensUsed": 1500,
       "operation": "llm_request",
       "model": "gpt-4"
     }
   }

**Example 3: Access Denial with Notification**

.. code-block:: json

   {
     "name": "DenyCustomerAccessAndSendWarningMongoDb",
     "inputs": {
       "customerId": "cust_12345",
       "reason": "Token limit exceeded",
       "currentUsage": 1050000,
       "limit": 1000000,
       "notifyEmail": true
     }
   }

API Key Management
------------------

**Key Structure**:

.. code-block:: text

   sk-proj-[environment]-[customer]-[random]
   
   Example: sk-proj-prod-acme-x7y8z9a0b1c2

**Key Properties**:

- Unique identifier
- Customer association
- Creation timestamp
- Expiration date
- Usage limits
- Allowed operations

User Management
---------------

**User Creation Flow**:

1. Check if user exists
2. Create new user if needed
3. Generate unique user ID
4. Set default permissions
5. Initialize usage counters

.. code-block:: json

   {
     "userId": "user_abc123",
     "customerId": "cust_12345",
     "created": "2024-01-15T10:00:00Z",
     "apiKeys": ["sk-proj-..."],
     "permissions": ["read", "write", "execute"],
     "tokenUsage": {
       "current": 0,
       "limit": 1000000,
       "resetDate": "2024-02-01T00:00:00Z"
     }
   }

Token Usage Tracking
--------------------

**Usage Metrics**:

- Tokens per request
- Cumulative daily usage
- Monthly aggregates
- Model-specific tracking
- Operation type breakdown

**Usage Report Example**:

.. code-block:: text

   Customer Usage Report - January 2024
   ------------------------------------
   Total Tokens: 850,000 / 1,000,000 (85%)
   
   By Model:
   - GPT-4: 600,000 tokens
   - GPT-3.5: 200,000 tokens
   - Embeddings: 50,000 tokens
   
   By Operation:
   - Chat Completions: 700,000
   - Embeddings: 50,000
   - Code Generation: 100,000

Access Control Rules
--------------------

**Rule Types**:

1. **Token Limits**: Hard and soft limits
2. **Rate Limiting**: Requests per minute/hour
3. **Operation Restrictions**: Allowed functions
4. **Time-based Access**: Business hours only
5. **IP Restrictions**: Whitelist/blacklist

**Rule Evaluation**:

.. code-block:: python

   def evaluate_access(request):
       # 1. Validate API key
       if not valid_key(request.api_key):
           return deny("Invalid API key")
       
       # 2. Check token limits
       if tokens_exceeded(request.customer_id):
           return deny("Token limit exceeded")
       
       # 3. Check rate limits
       if rate_exceeded(request.customer_id):
           return deny("Rate limit exceeded")
       
       # 4. Check permissions
       if not has_permission(request.user_id, request.operation):
           return deny("Insufficient permissions")
       
       return allow()

Notification System
-------------------

**Notification Triggers**:

- Token limit warnings (80%, 90%, 100%)
- Suspicious activity detection
- Failed authentication attempts
- Permission violations
- System security events

**Email Template Example**:

.. code-block:: text

   Subject: FlowKit Security Alert - Token Limit Exceeded
   
   Dear Customer,
   
   Your FlowKit account has exceeded its monthly token limit.
   
   Details:
   - Customer ID: cust_12345
   - Current Usage: 1,050,000 tokens
   - Monthly Limit: 1,000,000 tokens
   - Overage: 50,000 tokens
   
   Action Required:
   Please upgrade your plan or contact support.

Best Practices
--------------

1. **API Key Security**:
   - Rotate keys regularly
   - Use environment variables
   - Never commit keys to code
   - Implement key expiration

2. **Usage Monitoring**:
   - Set up usage alerts
   - Monitor unusual patterns
   - Track per-operation costs
   - Regular usage reviews

3. **Access Control**:
   - Principle of least privilege
   - Regular permission audits
   - Separate dev/prod keys
   - Document access policies

4. **Notification Management**:
   - Configure alert thresholds
   - Test notification delivery
   - Maintain contact lists
   - Log all notifications

Integration with Other Components
---------------------------------

Authentication integrates with:

- **All API Functions**: Pre-request validation
- **LLM Handler**: Token usage tracking
- **Logging System**: Security audit trails
- **Monitoring**: Real-time access metrics

Security Considerations
-----------------------

**Data Protection**:
- Encrypt API keys at rest
- Use secure key transmission
- Implement key hashing
- Regular security audits

**Attack Prevention**:
- Rate limiting
- IP blocking for abuse
- Anomaly detection
- Failed attempt tracking

Troubleshooting
---------------

**Authentication Failures**:
   - Verify API key format
   - Check key expiration
   - Confirm database connectivity
   - Review permission settings

**Usage Tracking Issues**:
   - Ensure counters are initialized
   - Check reset schedules
   - Verify aggregation logic
   - Monitor database performance

**Notification Failures**:
   - Test email configuration
   - Check template rendering
   - Verify recipient addresses
   - Review SMTP logs
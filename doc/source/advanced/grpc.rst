.. _grpc:

Running in production
=====================

Key considerations for production deployments.

Security basics
---------------

**Use an API key**
   Prevent unauthorized access

**Enable SSL**
   Encrypt data in transit

**Monitor access**
   Know who's using your system

Scaling up
----------

When you need more capacity:

- Run multiple instances
- Use a load balancer
- Monitor performance

Deployment checklist
--------------------

Before going live:

☐ Set a strong API key
☐ Enable SSL certificates
☐ Configure logging
☐ Test your workflows
☐ Plan for monitoring

Common deployment options
-------------------------

- **Docker** - Containerized deployment
- **Kubernetes** - Enterprise orchestration
- **Cloud services** - Managed infrastructure

Your infrastructure team can help choose the best option.

Next: :doc:`monitoring_logging`

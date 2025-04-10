---
sidebar_label: Using on Enterprise Grid
slug: /slack-cli/guides/using-slack-cli-on-an-enterprise-grid-organization
---

# Using Slack CLI on an Enterprise Grid organization

There are a few extra requirements, steps and considerations when you're developing on an Enterprise Grid organization.

While we'll discuss a few special privileges provided to those powerful Org Admins, the following material is recommended reading for all developers on an Enterprise Grid organization.

## Enterprise Grid terminology {#terminology}

This guide will use the following terms:
* Organization - The Enterprise Grid organization. The Slack CLI refers to your organization as your `team`. 
* Workspace - An Enterprise Grid organization can contain multiple workspaces. 
* Org Admins - An administrator of the Enterprise Grid organization. Org Primary Owners and Org Owners also have the privileges of Org Admins.
* User - Anyone with a Slack account in the Enterprise Grid organization.
* Developer - Someone building a workflow app in an Enterprise Grid organization, on any of its workspaces.

For example, imagine there is a _Kingdom of Britain organization_. _King Arthur_ is one of its Org Admins. This organization contains the _Knights of the Round Table_ workspace, of which _King Arthur_ is a member, along with other users such as _Mordred_ and _Lancelot_. 

With the terms defined, here's what Org Admins and developers need to know about the Slack CLI.

## Using a compatible version of the Slack CLI {#using}

Any developer using the Slack CLI in an organization will need to use v2.9.0 or later of Slack CLI.

Keeping your version of the Slack CLI up-to-date is recommended for a bevy of reasons. The Slack CLI will also prompt you to update.

If you're not sure if you're on a recent enough version, you can check with the following command:

```
slack --version
```

If you're experiencing login errors, try checking what version of the Slack CLI you're using! You won't be able to log in to an organization on earlier versions of the Slack CLI. 

## Logging in to an Enterprise Grid organization {#log-in}

Developers log in with the Slack CLI tool to **their entire organization**.

The login steps are the same as any other user, whether you are an Org Admin or not. Use the `slack login` command and paste the  `/slackauthticket <your auth ticket>` in any Slack channel or DM. Doing so will log you into the entire organization.

You can confirm this is the case with the `slack auth list` command. You'll get the following output:

```
OrgName (Team ID: E123ABC456)
User ID: U123ABC456
Last Updated: 2023-12-31 23:59:59 -7:00
Authorization Level: Organization
```

Since you're logged in to your organization, the apps you create will belong to your organization. 

### Enterprise Grid developers already logged into a workspace {#log-in-workspace}
Prior to Slack CLI v2.9.0, the Slack CLI was not compatible with an entire Enterprise Grid. Developers could, however, log in to a workspace of their organization. A developer who was developing with the Slack CLI prior to v2.9.0 may still show the following output from `slack auth list`:

```
WorkspaceName (Team ID: T123ABC456)
User ID: U123ABC456
Last Updated: 2023-12-31 23:59:59 -7:00
Authorization Level: Workspace
```

This legacy workspace-specific level authorization will still work, but is no longer grantable to new developers. If you log out or invalidate your login, you'll log in again with an organization level authorization.

Apps created with a legacy workspace-level authorization can still be managed by app owners or app collaborators with organization-level authorization, provided the workspace belongs to the organization.

## Requesting Admin Approval {#requesting-approval}

By default, [Admin Approval of Apps](/deno-slack-sdk/guides/controlling-permissions-for-admins) is turned on for all Slack Enterprise Grid organizations. 

Here's what the request flow looks like when trying to install an app to an organization:

1. Attempt to install your app with the CLI. The CLI will prompt you to request access if it is required. A request is sent to your organization admin. The request includes its [OAuth scopes](https://docs.slack.dev/authentication/) and outgoing domains.
2. Slackbot messages you confirming the request has been sent.
3. Slackbot messages the organization admin with the request.
4. An Org Admin approves the request.
5. Slackbot messages you informing you of the approval.
6. Attempt to install your app again. Installation should succeed. 

If your organization turned Admin Approval of Apps off, then no approval is needed. The first attempt to install your app should succeed.

## Granting access to specific workspaces {#granting-access}

While workflow apps are _installed_ on an entire organization, they do not necessarily have _access_ to every workspace in the organization.

Granting access to workspaces is done when an app is [run on a local development server](/deno-slack-sdk/guides/developing-locally) or [deployed to production](/deno-slack-sdk/guides/deploying-to-slack). The Slack CLI will prompt you to select whether to install with access to a single workspace or all workspaces in an organization. 

You can also skip the prompt and proactively grant the desired access by appending `-org-workspace-grant <team_id>` to the relevant command. 

For running on a local development server:

```
slack run -org-workspace-grant <team_id>
```

For deploying to production:

```
slack deploy -org-workspace-grant <team_id>
```

## Adjusting access to specific workspaces {#adjusting-access}

Org Admins can adjust an app's access grants anytime after granting their approval. Admins can change what workspaces the app has access to via Administration Settings. From there:

1. _Manage Organization_
2. _Integrations_
3. _Installed Apps_
4. _App Page_
5. _Manage_
6. Select _Remove from a workspace_ or _Add to more workspaces_

Use this process to ensure only desired apps are available in the proper workspaces.
create index issues_type_report on issues(crawl_id, issue_type_id, pagereport_id);
create index issue_type_priority on issue_types(priority);
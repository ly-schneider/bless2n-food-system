-- ownership
SELECT n.nspname, r.rolname AS owner
FROM pg_namespace n
JOIN pg_roles r ON r.oid = n.nspowner
WHERE n.nspname = 'public';

-- memberships
SELECT member.rolname AS member, role.rolname AS role
FROM pg_auth_members m
JOIN pg_roles role   ON role.oid = m.roleid
JOIN pg_roles member ON member.oid = m.member
WHERE role.rolname IN ('app_owner','app_runtime','app_readonly')
ORDER BY 1,2;
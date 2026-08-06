-- name: BillingActionItems :many
SELECT 'unpaid_claim' AS action_type,
       cl.id AS ref_id,
       CONCAT('Claim #', cl.id, ' for patient ', p.ptlname) AS description,
       cl.cltimestamp AS action_date,
       pr.procbalcurrent AS amount
FROM claimlog cl
LEFT JOIN procrec pr ON cl.clprocedure = pr.id
LEFT JOIN patient p ON pr.procpatient = p.id
WHERE pr.procbalcurrent > 0 AND pr.active = 'active'
UNION ALL
SELECT 'expiring_auth' AS action_type,
       a.id AS ref_id,
       CONCAT('Authorization #', a.id, ' for patient ', p.ptlname, ' expires ', DATE(a.authdtend)) AS description,
       a.authdtend AS action_date,
       0 AS amount
FROM authorizations a
LEFT JOIN patient p ON a.authpatient = p.id
WHERE a.authdtend BETWEEN CURDATE() AND DATE_ADD(CURDATE(), INTERVAL 30 DAY) AND a.active = 'active'
UNION ALL
SELECT 'unbilled_procedure' AS action_type,
       pr.id AS ref_id,
       CONCAT('Procedure ', c.cptnameext, ' for patient ', p.ptlname, ' on ', DATE(pr.procdt)) AS description,
       pr.procdt AS action_date,
       pr.proccharges AS amount
FROM procrec pr
LEFT JOIN cpt c ON pr.proccpt = c.id
LEFT JOIN patient p ON pr.procpatient = p.id
WHERE pr.procbalcurrent = pr.proccharges AND pr.proccharges > 0 AND pr.active = 'active'
ORDER BY action_date DESC
LIMIT 50;

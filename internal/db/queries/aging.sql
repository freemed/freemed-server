-- name: AgingSummary :many
SELECT
  SUM(procbalcurrent) AS balance,
  CASE
    WHEN DATEDIFF(NOW(), procdt) < 30 THEN '0-30'
    WHEN DATEDIFF(NOW(), procdt) < 60 THEN '31-60'
    WHEN DATEDIFF(NOW(), procdt) < 90 THEN '61-90'
    ELSE '90+'
  END AS age_bucket
FROM procrec
WHERE procdt IS NOT NULL
  AND procbalcurrent > 0
  AND active = 'active'
GROUP BY age_bucket;

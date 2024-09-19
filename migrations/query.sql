-- запрос на поиск всех тендеров
SELECT t.id, tc.name, tc.description, tc.type, t.status, organization_id, tc.version, t.created_at
FROM tender t
LEFT JOIN tender_condition tc on t.id = tc.tender_id
WHERE t.status = 'Published'
  AND tc.version = (
    SELECT MAX(tc2.version)
    FROM tender_condition tc2
    WHERE tc2.tender_id = t.id
)
ORDER BY tc.name
;

-- запрос на поиск username
SELECT *
FROM employee
WHERE username = 'asmith';

-- запрос на поиск права
SELECT id
FROM organization_responsible
WHERE organization_id = '77777777-7777-7777-7777-777777777777' AND user_id = '88888888-8888-8888-8888-888888888888';




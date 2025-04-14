INSERT INTO project (name, organization_id, description, industry,customer_persona, engagement_goals)
VALUES (:name, organization_id,:description,industry,customer_persona,engagement_goals) RETURNING id;

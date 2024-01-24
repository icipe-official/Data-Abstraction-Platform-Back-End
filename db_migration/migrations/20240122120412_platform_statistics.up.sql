-- platform statistics view
CREATE VIEW public.platform_statistics
 AS
SELECT
	(SELECT COUNT ('*') FROM public.projects) AS "no_of_projects",
	(SELECT COUNT ('*') FROM public.model_templates) AS "no_of_model_templates",
	(SELECT COUNT ('*') FROM public.catalogue) AS "no_of_catalogues",
	(SELECT COUNT ('*') FROM public.abstractions) AS "no_of_abstractions";

ALTER TABLE public.platform_statistics
    OWNER TO pg_database_owner;


-- project statistics view
CREATE VIEW public.project_statistics
 AS
SELECT
	projects.id AS "project_id",
	directory_projects_roles."no_of_members",
	model_templates."no_of_model_templates",
	catalogue."no_of_catalogues",
	files."no_of_files",
	abstractions."no_of_abstractions"
FROM 
	public.projects
	LEFT JOIN (
		SELECT
			directory_projects_roles.project_id AS "directory_projects_roles.project_id",
			COUNT (DISTINCT directory_projects_roles.directory_id) AS "no_of_members"
		FROM public.directory_projects_roles
			GROUP BY "directory_projects_roles.project_id"
	) AS directory_projects_roles ON "directory_projects_roles.project_id" = projects.id
	LEFT JOIN (
		SELECT
			model_templates.project_id AS "model_templates.project_id",
			COUNT ('*') AS "no_of_model_templates"
		FROM public.model_templates
			GROUP BY "model_templates.project_id"
	) AS model_templates ON "model_templates.project_id" = projects.id
	LEFT JOIN (
		SELECT
			catalogue.project_id AS "catalogue.project_id",
			COUNT ('*') AS "no_of_catalogues"
		FROM public.catalogue
			GROUP BY "catalogue.project_id"
	) AS catalogue ON "catalogue.project_id" = projects.id
	LEFT JOIN (
		SELECT
			files.project_id AS "files.project_id",
			COUNT ('*') AS "no_of_files"
		FROM public.files
			GROUP BY "files.project_id"
	) AS files ON "files.project_id" = projects.id
	LEFT JOIN (
		SELECT
			abstractions.project_id AS "abstractions.project_id",
			COUNT ('*') AS "no_of_abstractions"
		FROM public.abstractions
			GROUP BY "abstractions.project_id"
	) AS abstractions ON "abstractions.project_id" = projects.id;

ALTER TABLE public.project_statistics
    OWNER TO pg_database_owner;
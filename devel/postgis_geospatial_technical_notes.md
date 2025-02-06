## Postgres PostGiS Extension for Geospatial Needs - Technical Notes

Here some notes I gathered while implementing the initial version of query quotes around a particular radius. I tried to leave the notes in some kind of logical order that make sense.

- It appears there is two important geospatial type in PostGiS: `geography` and `geometry`. In a nutshell, in `geography` spherical coordinates are used in the various distance related mathematics to account for the Earth curvature. While `geometry` uses Cartesian coordinates, e.g. straight lines which are faster to compute at the expense of being approximation of distance over long distances.

 We decided to use `geography` to ensure better accuracy on terms of distance queries.

- Here a quick series of SQL I created around `quotes` table for testing purposes:

  ```sql
  select ST_AsEWKB('SRID=4326;POINT(-88.39 33.39)');
  select
    ST_X(ST_GeomFromEWKB(decode('0101000020E6100000A36000C0F51856C046C10080EBB14040', 'hex'))) as longitude,
    ST_Y(ST_GeomFromEWKB(decode('0101000020E6100000A36000C0F51856C046C10080EBB14040', 'hex'))) as latitude;

  With
    cte(id, origin, destination) as
    (VALUES
      ('a', 'SRID=4326;POINT(-79.995888 40.440624)'::geography, 'SRID=4326;POINT(-79.995888 40.440624)'::geography),
      ('b', 'SRID=4326;POINT(-80.059888 40.450624)'::geography, 'SRID=4326;POINT(-79.996888 40.441624)'::geography)
    ),
    points(point) as
      (VALUES
        ('SRID=4326;POINT(-80.081886 40.440624)'::geography)
      )
  select
    id,
    ST_X(origin::geometry) as origin_long,
    ST_Y(origin::geometry) as origin_lat,
    ST_X(destination::geometry) as destination_long,
    ST_Y(destination::geometry) as destination_lat,
    ST_Distance(origin, destination) / 1000 as distance_km,
    ST_Distance((select point from points), origin) / 1000 as point_to_origin_km,
    ST_Distance((select point from points), destination) / 1000 as point_to_destination_km
  from cte
  where
    ST_DWithin(origin, (select point from points)::geography, 30000) AND
    ST_DWithin(origin, (select point from points)::geography, 30000)
  order by point_to_origin_km asc;
  ```

### Reference

- `geography` reference manual https://postgis.net/workshops/postgis-intro/geography.html
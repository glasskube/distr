select now();

select deployment_id,
       created_at,
       created_at - lag(created_at, 1, localtimestamp - interval '24 hour') over (
         partition by deployment_id order by created_at) as diff_to_prev
from deploymentstatus
where created_at > localtimestamp - interval '24 hour'
order by deployment_id, created_at;

explain
select
  d.deployment_id,
  extract(epoch from interval '24 hour') / extract(epoch from interval '2 minute') as expected_status_count,
  floor(extract(epoch from d.diff_to_prev) / extract(epoch from interval '2 minute'))::int as not_received_status_count
from (
       select deployment_id,
              created_at,
              created_at - lag(created_at, 1, localtimestamp - interval '24 hour') over (
                partition by deployment_id order by created_at) as diff_to_prev
       from deploymentstatus
       where created_at > localtimestamp - interval '24 hour'
       order by deployment_id, created_at
     ) as d;

-- 24 hours total
select
  d.deployment_id,
  extract(epoch from interval '24 hour') / extract(epoch from interval '10 second') as totalIntervals,
  sum(floor(extract(epoch from d.diff_to_prev) / extract(epoch from interval '10 second'))::int) +
  floor(extract(epoch from now() - max(d.created_at)) / extract(epoch from interval '10 second'))::int
                                                                                    as timesStatusNotReceived
from (
       select deployment_id,
              created_at,
              created_at - lag(created_at, 1, now() - interval '24 hour') over (
                partition by deployment_id order by created_at) as diff_to_prev
       from deploymentstatus
       where created_at > now() - interval '24 hour'
       order by deployment_id, created_at
     ) as d
where d.deployment_id = '11a676ea-714a-4e5e-9a67-54c9d2f7b64c'
group by d.deployment_id;

select
  date_trunc('hour', d.created_at) as hour,
  -- d.deployment_id,
  extract(epoch from interval '1 hour') / extract(epoch from interval '10 second') as totalIntervals,
  sum(floor(extract(epoch from d.diff_to_prev) / extract(epoch from interval '10 second'))::int) +
  (CASE WHEN (date_trunc('hour', now()) = date_trunc('hour', max(d.created_at))) THEN (
    floor(extract(epoch from now() - max(d.created_at)) / extract(epoch from interval '10 second'))::int
    ) ELSE 0 END) as timesStatusNotReceived
from (
       select deployment_id,
              created_at,
              created_at - lag(created_at) over (
                partition by deployment_id order by created_at) as diff_to_prev
       from deploymentstatus
       where created_at > now() - interval '24 hour'
       order by deployment_id, created_at
     ) as d
where d.deployment_id = '5d5e4e61-cd82-46ed-965b-6ba5d3a1d1b8'
group by hour, d.deployment_id
order by 1;

SELECT date_trunc('hour', x)
FROM   generate_series(now() - interval '23 hour', now(), interval  '1 hour') AS x;

-- one specific hour
select
  d.deployment_id,
  extract(epoch from interval '1 hour') / extract(epoch from interval '10 second') as totalIntervals,
  sum(floor(extract(epoch from d.diff_to_prev) / extract(epoch from interval '10 second'))::int) +
  floor(extract(epoch from '2024-12-18 12:00:00' - max(d.created_at)) / extract(epoch from interval '10 second'))::int
                                                                                   as timesStatusNotReceived
from (
       select deployment_id,
              created_at,
              created_at - lag(created_at, 1, '2024-12-18 11:00:00') over (
                partition by deployment_id order by created_at) as diff_to_prev
       from deploymentstatus
       where created_at between '2024-12-18 11:00:00' and '2024-12-18 12:00:00'
       order by deployment_id, created_at
     ) as d
where d.deployment_id = '11a676ea-714a-4e5e-9a67-54c9d2f7b64c'
group by d.deployment_id;




select deployment_id,
       created_at,
       created_at - lag(created_at) over (
         partition by deployment_id order by created_at) as diff_to_prev
from deploymentstatus
where created_at > localtimestamp - interval '24 hour'
order by deployment_id, created_at;

select *, now()
from deploymentstatus
where deployment_id = '11a676ea-714a-4e5e-9a67-54c9d2f7b64c'
order by created_at desc
limit 1;

explain
select *
from deploymentstatus
where created_at > localtimestamp - interval '24 hour'
order by deployment_id, created_at;

create index status_test ON deploymentstatus(deployment_id, created_at ASC);

select count(*) from deploymentstatus;


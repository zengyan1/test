select disable_fault_injection();
select enable_fault_injection();
select add_fault_point('ret', ':::', 'return', 0, '');
select trigger_fault_point('ret');
select trigger_fault_point('ret');
select trigger_fault_point('ret');
select add_fault_point('cnt', ':::', 'getcount', 0, 'ret');
select trigger_fault_point('cnt');
-- select add_fault_point('panic', ':::', 'panic', 0, '');
-- select trigger_fault_point('panic');
select disable_fault_injection();

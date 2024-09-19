INSERT INTO employee (id, username, first_name, last_name, created_at, updated_at)
VALUES
    ('22222222-2222-2222-2222-222222222222','jdoe', 'John', 'Doe', '2006-01-02T15:04:05Z07:00', '2006-01-02T15:04:05Z07:00'),
    ('44444444-4444-4444-4444-444444444444','asmith', 'Alice', 'Smith', '2007-01-02T15:04:05Z07:00', '2008-01-02T15:04:05Z07:00'),
    ('66666666-6666-6666-6666-666666666666','bwilliams', 'Bob', 'Williams', '2008-01-02T15:04:05Z07:00', '2009-01-02T15:04:05Z07:00'),
    ('88888888-8888-8888-8888-888888888888', 'cjones', 'Carol', 'Jones', '2008-01-02T15:04:05Z07:00', '2008-01-02T15:04:05Z07:00'),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','djohnson', 'David', 'Johnson', '2009-01-02T15:04:05Z07:00', '2009-01-02T15:04:05Z07:00')
    ON CONFLICT (id) DO NOTHING;

INSERT INTO organization (id, name, description, type, created_at, updated_at)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'Alpha Corp', 'Leading company in manufacturing', 'LLC', '2007-01-02T15:04:05Z07:00', '2007-01-02T15:04:05Z07:00'),
    ('33333333-3333-3333-3333-333333333333', 'Beta LLC', 'Construction and architecture', 'LLC', '2006-01-02T15:04:05Z07:00', '2006-01-02T15:04:05Z07:00'),
    ('55555555-5555-5555-5555-555555555555', 'Gamma Inc', 'Logistics and delivery services', 'JSC', '2007-01-02T15:04:05Z07:00', '2010-01-02T15:04:05Z07:00'),
    ('77777777-7777-7777-7777-777777777777', 'Delta IE', 'Independent entrepreneur in construction', 'IE', '2005-01-02T15:04:05Z07:00', '2005-01-02T15:04:05Z07:00'),
    ('99999999-9999-9999-9999-999999999999', 'Epsilon LLC', 'Software development company', 'LLC', '2011-01-02T15:04:05Z07:00', '2012-01-02T15:04:05Z07:00')
    ON CONFLICT (id) DO NOTHING;

INSERT INTO organization_responsible (id, organization_id, user_id)
VALUES
    ('00000000-3333-3333-3333-333333333333', '55555555-5555-5555-5555-555555555555', '66666666-6666-6666-6666-666666666666'),
    ('00000000-4444-4444-4444-444444444444', '77777777-7777-7777-7777-777777777777', '88888888-8888-8888-8888-888888888888'),
    ('00000000-5555-5555-5555-555555555555', '99999999-9999-9999-9999-999999999999', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'),
    ('00000000-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', '22222222-2222-2222-2222-222222222222'),
    ('00000000-2222-2222-2222-222222222222', '33333333-3333-3333-3333-333333333333', '44444444-4444-4444-4444-444444444444')
    ON CONFLICT (id) DO NOTHING;


INSERT INTO tender (id, status, organization_id, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-111111111111', 'Created', '11111111-1111-1111-1111-111111111111', '2012-01-02T15:04:05Z07:00', '2012-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-333333333333', 'Published',  '33333333-3333-3333-3333-333333333333', '2013-01-02T15:04:05Z07:00', '2014-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-777777777777', 'Closed', '55555555-5555-5555-5555-555555555555', '2014-01-02T15:04:05Z07:00', '2015-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-555555555555', 'Created', '77777777-7777-7777-7777-777777777777', '2015-01-02T15:04:05Z07:00', '2015-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-999999999999', 'Published', '99999999-9999-9999-9999-999999999999', '2016-01-02T15:04:05Z07:00', '2016-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-888888888888', 'Published', '99999999-9999-9999-9999-999999999999', '2016-01-02T15:04:05Z07:00', '2016-01-02T15:04:05Z07:00')
    ON CONFLICT (id) DO NOTHING;


INSERT INTO tender_condition (tender_id, name, description, type, version)
VALUES
    ('00000000-0000-0000-0000-111111111111', 'Condition 1', 'First condition for tender', 'Construction', 1),
    ('00000000-0000-0000-0000-333333333333', 'Condition 2', 'Second condition for tender', 'Delivery', 1),
    ('00000000-0000-0000-0000-333333333333', 'Condition 2.1', 'Second condition for tender', 'Delivery', 2),
    ('00000000-0000-0000-0000-777777777777', 'Z Condition 3', 'Third condition for tender', 'Manufacture', 1),
    ('00000000-0000-0000-0000-777777777777', 'Z Condition 3.1', 'Third condition for tender', 'Manufacture', 2),
    ('00000000-0000-0000-0000-777777777777', 'Z Condition 3.2', 'Third condition for tender', 'Manufacture', 3),
    ('00000000-0000-0000-0000-555555555555', 'Condition 4', 'Fourth condition for tender', 'Construction', 1),
    ('00000000-0000-0000-0000-999999999999', 'X Condition 5', 'Fifth condition for tender', 'Delivery', 1),
    ('00000000-0000-0000-0000-999999999999', 'X Condition 5.1', 'Fifth condition for tender', 'Delivery', 2),
    ('00000000-0000-0000-0000-888888888888', 'A Condition 6', 'Six condition for tender', 'Construction', 1),
    ('00000000-0000-0000-0000-888888888888', 'A Condition 6.1', 'Six condition for tender', 'Construction', 2),
    ('00000000-0000-0000-0000-888888888888', 'A Condition 6.2', 'Six condition for tender', 'Construction', 3),
    ('00000000-0000-0000-0000-999999999999', 'X Condition 5.2', 'Fifth condition for tender', 'Delivery', 3),
    ('00000000-0000-0000-0000-888888888888', 'A Condition 6.4', 'Six condition for tender', 'Construction', 4)
    ON CONFLICT (tender_id, version) DO NOTHING;

INSERT INTO bid (id, status, tender_id, author_type, author_id, created_at, updated_at)
VALUES
    ('00000000-0000-0000-0000-444444444444', 'Created', '00000000-0000-0000-0000-111111111111', 'User', '44444444-4444-4444-4444-444444444444', '2012-01-02T15:04:05Z07:00', '2012-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-444444444441', 'Published', '00000000-0000-0000-0000-111111111111', 'User', '66666666-6666-6666-6666-666666666666', '2012-01-02T15:04:05Z07:00', '2012-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-000000000000', 'Published', '00000000-0000-0000-0000-333333333333', 'User', '22222222-2222-2222-2222-222222222222', '2012-01-02T15:04:05Z07:00', '2012-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-222222222222', 'Canceled', '00000000-0000-0000-0000-777777777777', 'User', '66666666-6666-6666-6666-666666666666', '2012-01-02T15:04:05Z07:00', '2012-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-888888888888', 'Created',  '00000000-0000-0000-0000-555555555555', 'User', '88888888-8888-8888-8888-888888888888', '2012-01-02T15:04:05Z07:00', '2012-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-0000-666666666666', 'Created', '00000000-0000-0000-0000-999999999999', 'User', '22222222-2222-2222-2222-222222222222', '2012-01-02T15:04:05Z07:00', '2012-01-02T15:04:05Z07:00'),
    ('00000000-0000-0000-8888-888888888881', 'Published', '00000000-0000-0000-0000-555555555555', 'User', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' , '2012-01-02T15:04:05Z07:00', '2012-01-02T15:04:05Z07:00')
    ON CONFLICT (id) DO NOTHING;


INSERT INTO bid_condition (bid_id, name, description, version)
VALUES
    ('00000000-0000-0000-0000-444444444444', 'Bid Condition 1', 'First condition for bid', 1),
    ('00000000-0000-0000-0000-000000000000', 'Bid Condition 2', 'Second condition for bid', 1),
    ('00000000-0000-0000-0000-000000000000', 'Bid Condition 2.2', 'Second condition for bid', 2),
    ('00000000-0000-0000-0000-222222222222', 'Bid Condition 3', 'Third condition for bid', 1),
    ('00000000-0000-0000-0000-888888888888', 'Bid Condition 4', 'Fourth condition for bid', 1),
    ('00000000-0000-0000-0000-888888888888', 'Bid Condition 4.2', 'Fourth condition for bid', 2),
    ('00000000-0000-0000-0000-888888888888', 'Bid Condition 4.3', 'Fourth condition for bid', 3),
    ('00000000-0000-0000-0000-666666666666', 'Bid Condition 5', 'Fifth condition for bid', 1),
    ('00000000-0000-0000-8888-888888888881', 'Bid Condition 1', 'First condition for bid', 1),
    ('00000000-0000-0000-0000-444444444441', 'Bid Condition 1', 'First condition for bid', 1),
    ('00000000-0000-0000-0000-444444444441', 'Bid Condition 1.2', 'First condition for bid', 2),
    ('00000000-0000-0000-8888-888888888881', 'Bid Condition 1.2', 'First condition for bid', 2)
    ON CONFLICT (bid_id, version) DO NOTHING;

INSERT INTO bid_feedback (bid_id, feedback, username)
VALUES
    ('00000000-0000-0000-0000-444444444444', 'Great bid!', 'jdoe'),
    ('00000000-0000-0000-0000-000000000000', 'Needs improvement.', 'asmith'),
    ('00000000-0000-0000-0000-222222222222', 'Perfect conditions.', 'bwilliams'),
    ('00000000-0000-0000-0000-888888888888', 'Satisfied with the proposal.', 'cjones'),
    ('00000000-0000-0000-0000-666666666666', 'Bid rejected.', 'djohnson');
-- Delete the three default admins created for local development.
-- Change the emails if different seed emails were used.
DELETE FROM admins
WHERE email IN (
    'fdlly183@gmail.com',
    'haerunnas.sg@gmail.com',
    'febrianbenedict@gmail.com'
);

#! /usr/local/bin/bash

echo -n "acct pw for stripe_acct: " >&2
read -s pw


cat <<RUN_PRIV
#use strip_proj;

grant all on strip_proj.* to 'stripe_acct';
create user stripe_acct identified by '$pw';

RUN_PRIV


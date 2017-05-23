#!/usr/bin/env perl

use v5.12;
use Statistics::Basic qw(:all);
use Data::Dumper;

$| = 1; #turn on autoflush
#my @VEC;
my $s = {'32'=>{'full'=>{'get'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'put'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'del'=>{'data'=>[],'mean'=>0,'stddev'=>0}},
                'comp'=>{'get'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'put'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'del'=>{'data'=>[],'mean'=>0,'stddev'=>0}},
                'hybr'=>{'get'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'put'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'del'=>{'data'=>[],'mean'=>0,'stddev'=>0}}
               },
         '64'=>{'full'=>{'get'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'put'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'del'=>{'data'=>[],'mean'=>0,'stddev'=>0}},
                'comp'=>{'get'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'put'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'del'=>{'data'=>[],'mean'=>0,'stddev'=>0}},
                'hybr'=>{'get'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'put'=>{'data'=>[],'mean'=>0,'stddev'=>0},
                         'del'=>{'data'=>[],'mean'=>0,'stddev'=>0}}
               }
        };

my $ttype;
LINE: while (my $l = <>) {
	chomp $l;
	#say $l;
	# Set the current $ttype
	$l =~ m/type = fullonly/ && do {
		$ttype = 'full';
		#say "Found \$ttype = '$ttype'";
		next LINE;
	};
	$l =~ m/type = componly/ && do {
		$ttype = 'comp';
		#say "Found \$ttype = '$ttype'";
		next LINE;
	};
	$l =~ m/type = hybrid/ && do {
		$ttype = 'hybr';
		#say "Found \$ttype = '$ttype'";
		next LINE;
	};
	# record the benchmark result
	$l =~ m/^BenchmarkHamt32Get-8/ && do {
		my $bit = '32';
		my $op = 'get';
		my @fields = split(' ', $l);
		push @{$s->{$bit}{$ttype}{$op}{'data'}}, $fields[2];
		#say(join(" ", $bit, $ttype, $op, $fields[2]));
		local $Data::Dumper::Indent = 0;
		#say Dumper( $s->{$bit}{$ttype}{$op}{'data'} );
		next LINE;
	};
	$l =~ m/^BenchmarkHamt32Put-8/ && do {
		my $bit = '32';
		my $op = 'put';
		my @fields = split(' ', $l);
		push @{$s->{$bit}{$ttype}{$op}{'data'}}, $fields[2];
		#say(join(" ", $bit, $ttype, $op, $fields[2]));
		local $Data::Dumper::Indent = 0;
		#say Dumper( $s->{$bit}{$ttype}{$op}{'data'} );
		next LINE;
	};
	$l =~ m/^BenchmarkHamt32Del-8/ && do {
		my $bit = '32';
		my $op = 'del';
		my @fields = split(' ', $l);
		push @{$s->{$bit}{$ttype}{$op}{'data'}}, $fields[2];
		#say(join(" ", $bit, $ttype, $op, $fields[2]));
		local $Data::Dumper::Indent = 0;
		#say Dumper( $s->{$bit}{$ttype}{$op}{'data'} );
		next LINE;
	};
	$l =~ m/^BenchmarkHamt64Get-8/ && do {
		my $bit = '64';
		my $op = 'get';
		my @fields = split(' ', $l);
		push @{$s->{$bit}{$ttype}{$op}{'data'}}, $fields[2];
		#say(join(" ", $bit, $ttype, $op, $fields[2]));
		local $Data::Dumper::Indent = 0;
		#say Dumper( $s->{$bit}{$ttype}{$op}{'data'} );
		next LINE;
	};
	$l =~ m/^BenchmarkHamt64Put-8/ && do {
		my $bit = '64';
		my $op = 'put';
		my @fields = split(' ', $l);
		push @{$s->{$bit}{$ttype}{$op}{'data'}}, $fields[2];
		#say(join(" ", $bit, $ttype, $op, $fields[2]));
		local $Data::Dumper::Indent = 0;
		#say Dumper( $s->{$bit}{$ttype}{$op}{'data'} );
		next LINE;
	};
	$l =~ m/^BenchmarkHamt64Del-8/ && do {
		my $bit = '64';
		my $op = 'del';
		my @fields = split(' ', $l);
		push @{$s->{$bit}{$ttype}{$op}{'data'}}, $fields[2];
		#say(join(" ", $bit, $ttype, $op, $fields[2]));
		local $Data::Dumper::Indent = 0;
		#say Dumper( $s->{$bit}{$ttype}{$op}{'data'} );
		next LINE;
	};
}

# Calculate mean & dev for each data vector
for my $bit (keys %$s) {
	for my $ttype (keys %{$s->{$bit}}) {
		for my $op (keys %{$s->{$bit}{$ttype}}) {
			my $entry = $s->{$bit}{$ttype}{$op};
			@{$entry->{'data'}} == 0 && next;
			my $v = vector(@{$entry->{'data'}});
			my $m = $entry->{'mean'} = mean($v);
			my $d = $entry->{'stddev'} = stddev($v);
			say "$bit/$ttype/$op => $m +/- $d ns";
		}
	}
}

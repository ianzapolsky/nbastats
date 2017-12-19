#!/usr/bin/env python

# this script downloads the list of game IDs for a specific season.

from statsnba.api import Api
import pandas as pd

import argparse
parser = argparse.ArgumentParser(description='Download Pbp')

parser.add_argument('-s', '--season', required=True,
                    default='', help='season to pull. e.g. 2015-16')

parser.add_argument('-o', '--output', required=True,
                    default='output', help='file to save')

parser.add_argument('-f', '--format', dest='format',
                    default='csv', choices={'csv', 'excel'}, action='store')

parser.add_argument('-t', '--type', required=False,
                    default='Regular Season', choices={'Regular Season', 'Playoffs'}, help='season type: "Regular Season" or "Playoffs"')

if __name__ == '__main__':
    args = parser.parse_args()

    print('Downloading game ids for %s %s' % (args.season, args.type))

    api = Api()
    result = api.GetSeasonGameIDs(args.season, args.type)
    df = pd.DataFrame(result)

    print('Saving data to %s' % args.output)

    if args.format == 'csv':
        df.to_csv(args.output)
    else:
        df.to_excel(args.output)


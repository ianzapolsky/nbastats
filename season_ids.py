#!/usr/bin/env python

# example: 0020901003

from statsnba.api import Api
import pandas as pd

import argparse
parser = argparse.ArgumentParser(description='Download Pbp')

parser.add_argument('-s', '--season', required=True,
                    default='', help='season to pull. e.g. 2015-16')

parser.add_argument('-o', '--output', required=True,
                    default='output', help='file to save')

parser.add_argument('-f', '--format', dest='format',
                    default='excel', choices={'csv', 'excel'}, action='store')

if __name__ == '__main__':
    args = parser.parse_args()
    season_type = 'Playoffs'

    print 'Downloading game ids for season {0}'.format(args.season)

    api = Api()
    result = api.GetSeasonGameIDs(args.season, season_type)
    df = pd.DataFrame(result)

    print 'Saving to {0}'.format(args.output)

    if args.format == 'csv':
        df.to_csv(args.output)
    else:
        df.to_excel(args.output)


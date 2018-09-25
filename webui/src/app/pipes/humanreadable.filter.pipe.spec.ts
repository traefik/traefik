import { HumanReadableFilterPipe } from './humanreadable.filter.pipe';

describe('HumanReadableFilterPipe', () => {
    const pipe = new HumanReadableFilterPipe();

    const datatable = [{
        'given': '180000000000',
        'expected': '180s'
    },
    {
        'given': '4096.0',
        'expected': '4096ns'
    },
    {
        'given': '7200000000000',
        'expected': '120m'
    },
    {
        'given': '1337',
        'expected': '1337ns'
    },
    {
        'given': 'traefik',
        'expected': 'traefik',
    },
    {
        'given': '-23',
        'expected': '-23',
    },
    {
        'given': '0',
        'expected': '0',
    },
    ];

    datatable.forEach(item => {
        it((item.given + ' should be transformed to ' + item.expected ), () => {
            expect(pipe.transform(item.given)).toEqual(item.expected);
        });
    });

    it('create an instance', () => {
        expect(pipe).toBeTruthy();
    });

});

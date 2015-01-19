/* A Bison parser, made by GNU Bison 2.3.  */

/* Skeleton implementation for Bison's Yacc-like parsers in C

   Copyright (C) 1984, 1989, 1990, 2000, 2001, 2002, 2003, 2004, 2005, 2006
   Free Software Foundation, Inc.

   This program is free software; you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation; either version 2, or (at your option)
   any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program; if not, write to the Free Software
   Foundation, Inc., 51 Franklin Street, Fifth Floor,
   Boston, MA 02110-1301, USA.  */

/* As a special exception, you may create a larger work that contains
   part or all of the Bison parser skeleton and distribute that work
   under terms of your choice, so long as that work isn't itself a
   parser generator using the skeleton or a modified version thereof
   as a parser skeleton.  Alternatively, if you modify or redistribute
   the parser skeleton itself, you may (at your option) remove this
   special exception, which will cause the skeleton and the resulting
   Bison output files to be licensed under the GNU General Public
   License without this special exception.

   This special exception was added by the Free Software Foundation in
   version 2.2 of Bison.  */

/* C LALR(1) parser skeleton written by Richard Stallman, by
   simplifying the original so-called "semantic" parser.  */

/* All symbols defined below should begin with yy or YY, to avoid
   infringing on user name space.  This should be done even for local
   variables, as they might otherwise be expanded by user macros.
   There are some unavoidable exceptions within include files to
   define necessary library symbols; they are noted "INFRINGES ON
   USER NAME SPACE" below.  */

/* Identify Bison output.  */
#define YYBISON 1

/* Bison version.  */
#define YYBISON_VERSION "2.3"

/* Skeleton name.  */
#define YYSKELETON_NAME "yacc.c"

/* Pure parsers.  */
#define YYPURE 0

/* Using locations.  */
#define YYLSP_NEEDED 0



/* Tokens.  */
#ifndef YYTOKENTYPE
# define YYTOKENTYPE
   /* Put the tokens into the symbol table, so that GDB and other debuggers
      know about them.  */
   enum yytokentype {
     LLITERAL = 258,
     LASOP = 259,
     LCOLAS = 260,
     LBREAK = 261,
     LCASE = 262,
     LCHAN = 263,
     LCONST = 264,
     LCONTINUE = 265,
     LDDD = 266,
     LDEFAULT = 267,
     LDEFER = 268,
     LELSE = 269,
     LFALL = 270,
     LFOR = 271,
     LFUNC = 272,
     LGO = 273,
     LGOTO = 274,
     LIF = 275,
     LIMPORT = 276,
     LINTERFACE = 277,
     LMAP = 278,
     LNAME = 279,
     LPACKAGE = 280,
     LRANGE = 281,
     LRETURN = 282,
     LSELECT = 283,
     LSTRUCT = 284,
     LSWITCH = 285,
     LTYPE = 286,
     LVAR = 287,
     LANDAND = 288,
     LANDNOT = 289,
     LBODY = 290,
     LCOMM = 291,
     LDEC = 292,
     LEQ = 293,
     LGE = 294,
     LGT = 295,
     LIGNORE = 296,
     LINC = 297,
     LLE = 298,
     LLSH = 299,
     LLT = 300,
     LNE = 301,
     LOROR = 302,
     LRSH = 303,
     NotPackage = 304,
     NotParen = 305,
     PreferToRightParen = 306
   };
#endif
/* Tokens.  */
#define LLITERAL 258
#define LASOP 259
#define LCOLAS 260
#define LBREAK 261
#define LCASE 262
#define LCHAN 263
#define LCONST 264
#define LCONTINUE 265
#define LDDD 266
#define LDEFAULT 267
#define LDEFER 268
#define LELSE 269
#define LFALL 270
#define LFOR 271
#define LFUNC 272
#define LGO 273
#define LGOTO 274
#define LIF 275
#define LIMPORT 276
#define LINTERFACE 277
#define LMAP 278
#define LNAME 279
#define LPACKAGE 280
#define LRANGE 281
#define LRETURN 282
#define LSELECT 283
#define LSTRUCT 284
#define LSWITCH 285
#define LTYPE 286
#define LVAR 287
#define LANDAND 288
#define LANDNOT 289
#define LBODY 290
#define LCOMM 291
#define LDEC 292
#define LEQ 293
#define LGE 294
#define LGT 295
#define LIGNORE 296
#define LINC 297
#define LLE 298
#define LLSH 299
#define LLT 300
#define LNE 301
#define LOROR 302
#define LRSH 303
#define NotPackage 304
#define NotParen 305
#define PreferToRightParen 306




/* Copy the first part of user declarations.  */
#line 20 "go.y"

#include <u.h>
#include <stdio.h>	/* if we don't, bison will, and go.h re-#defines getc */
#include <libc.h>
#include "go.h"

static void fixlbrace(int);


/* Enabling traces.  */
#ifndef YYDEBUG
# define YYDEBUG 0
#endif

/* Enabling verbose error messages.  */
#ifdef YYERROR_VERBOSE
# undef YYERROR_VERBOSE
# define YYERROR_VERBOSE 1
#else
# define YYERROR_VERBOSE 1
#endif

/* Enabling the token table.  */
#ifndef YYTOKEN_TABLE
# define YYTOKEN_TABLE 0
#endif

#if ! defined YYSTYPE && ! defined YYSTYPE_IS_DECLARED
typedef union YYSTYPE
#line 28 "go.y"
{
	Node*		node;
	NodeList*		list;
	Type*		type;
	Sym*		sym;
	struct	Val	val;
	int		i;
}
/* Line 193 of yacc.c.  */
#line 216 "y.tab.c"
	YYSTYPE;
# define yystype YYSTYPE /* obsolescent; will be withdrawn */
# define YYSTYPE_IS_DECLARED 1
# define YYSTYPE_IS_TRIVIAL 1
#endif



/* Copy the second part of user declarations.  */


/* Line 216 of yacc.c.  */
#line 229 "y.tab.c"

#ifdef short
# undef short
#endif

#ifdef YYTYPE_UINT8
typedef YYTYPE_UINT8 yytype_uint8;
#else
typedef unsigned char yytype_uint8;
#endif

#ifdef YYTYPE_INT8
typedef YYTYPE_INT8 yytype_int8;
#elif (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
typedef signed char yytype_int8;
#else
typedef short int yytype_int8;
#endif

#ifdef YYTYPE_UINT16
typedef YYTYPE_UINT16 yytype_uint16;
#else
typedef unsigned short int yytype_uint16;
#endif

#ifdef YYTYPE_INT16
typedef YYTYPE_INT16 yytype_int16;
#else
typedef short int yytype_int16;
#endif

#ifndef YYSIZE_T
# ifdef __SIZE_TYPE__
#  define YYSIZE_T __SIZE_TYPE__
# elif defined size_t
#  define YYSIZE_T size_t
# elif ! defined YYSIZE_T && (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
#  include <stddef.h> /* INFRINGES ON USER NAME SPACE */
#  define YYSIZE_T size_t
# else
#  define YYSIZE_T unsigned int
# endif
#endif

#define YYSIZE_MAXIMUM ((YYSIZE_T) -1)

#ifndef YY_
# if defined YYENABLE_NLS && YYENABLE_NLS
#  if ENABLE_NLS
#   include <libintl.h> /* INFRINGES ON USER NAME SPACE */
#   define YY_(msgid) dgettext ("bison-runtime", msgid)
#  endif
# endif
# ifndef YY_
#  define YY_(msgid) msgid
# endif
#endif

/* Suppress unused-variable warnings by "using" E.  */
#if ! defined lint || defined __GNUC__
# define YYUSE(e) ((void) (e))
#else
# define YYUSE(e) /* empty */
#endif

/* Identity function, used to suppress warnings about constant conditions.  */
#ifndef lint
# define YYID(n) (n)
#else
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static int
YYID (int i)
#else
static int
YYID (i)
    int i;
#endif
{
  return i;
}
#endif

#if ! defined yyoverflow || YYERROR_VERBOSE

/* The parser invokes alloca or malloc; define the necessary symbols.  */

# ifdef YYSTACK_USE_ALLOCA
#  if YYSTACK_USE_ALLOCA
#   ifdef __GNUC__
#    define YYSTACK_ALLOC __builtin_alloca
#   elif defined __BUILTIN_VA_ARG_INCR
#    include <alloca.h> /* INFRINGES ON USER NAME SPACE */
#   elif defined _AIX
#    define YYSTACK_ALLOC __alloca
#   elif defined _MSC_VER
#    include <malloc.h> /* INFRINGES ON USER NAME SPACE */
#    define alloca _alloca
#   else
#    define YYSTACK_ALLOC alloca
#    if ! defined _ALLOCA_H && ! defined _STDLIB_H && (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
#     include <stdlib.h> /* INFRINGES ON USER NAME SPACE */
#     ifndef _STDLIB_H
#      define _STDLIB_H 1
#     endif
#    endif
#   endif
#  endif
# endif

# ifdef YYSTACK_ALLOC
   /* Pacify GCC's `empty if-body' warning.  */
#  define YYSTACK_FREE(Ptr) do { /* empty */; } while (YYID (0))
#  ifndef YYSTACK_ALLOC_MAXIMUM
    /* The OS might guarantee only one guard page at the bottom of the stack,
       and a page size can be as small as 4096 bytes.  So we cannot safely
       invoke alloca (N) if N exceeds 4096.  Use a slightly smaller number
       to allow for a few compiler-allocated temporary stack slots.  */
#   define YYSTACK_ALLOC_MAXIMUM 4032 /* reasonable circa 2006 */
#  endif
# else
#  define YYSTACK_ALLOC YYMALLOC
#  define YYSTACK_FREE YYFREE
#  ifndef YYSTACK_ALLOC_MAXIMUM
#   define YYSTACK_ALLOC_MAXIMUM YYSIZE_MAXIMUM
#  endif
#  if (defined __cplusplus && ! defined _STDLIB_H \
       && ! ((defined YYMALLOC || defined malloc) \
	     && (defined YYFREE || defined free)))
#   include <stdlib.h> /* INFRINGES ON USER NAME SPACE */
#   ifndef _STDLIB_H
#    define _STDLIB_H 1
#   endif
#  endif
#  ifndef YYMALLOC
#   define YYMALLOC malloc
#   if ! defined malloc && ! defined _STDLIB_H && (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
void *malloc (YYSIZE_T); /* INFRINGES ON USER NAME SPACE */
#   endif
#  endif
#  ifndef YYFREE
#   define YYFREE free
#   if ! defined free && ! defined _STDLIB_H && (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
void free (void *); /* INFRINGES ON USER NAME SPACE */
#   endif
#  endif
# endif
#endif /* ! defined yyoverflow || YYERROR_VERBOSE */


#if (! defined yyoverflow \
     && (! defined __cplusplus \
	 || (defined YYSTYPE_IS_TRIVIAL && YYSTYPE_IS_TRIVIAL)))

/* A type that is properly aligned for any stack member.  */
union yyalloc
{
  yytype_int16 yyss;
  YYSTYPE yyvs;
  };

/* The size of the maximum gap between one aligned stack and the next.  */
# define YYSTACK_GAP_MAXIMUM (sizeof (union yyalloc) - 1)

/* The size of an array large to enough to hold all stacks, each with
   N elements.  */
# define YYSTACK_BYTES(N) \
     ((N) * (sizeof (yytype_int16) + sizeof (YYSTYPE)) \
      + YYSTACK_GAP_MAXIMUM)

/* Copy COUNT objects from FROM to TO.  The source and destination do
   not overlap.  */
# ifndef YYCOPY
#  if defined __GNUC__ && 1 < __GNUC__
#   define YYCOPY(To, From, Count) \
      __builtin_memcpy (To, From, (Count) * sizeof (*(From)))
#  else
#   define YYCOPY(To, From, Count)		\
      do					\
	{					\
	  YYSIZE_T yyi;				\
	  for (yyi = 0; yyi < (Count); yyi++)	\
	    (To)[yyi] = (From)[yyi];		\
	}					\
      while (YYID (0))
#  endif
# endif

/* Relocate STACK from its old location to the new one.  The
   local variables YYSIZE and YYSTACKSIZE give the old and new number of
   elements in the stack, and YYPTR gives the new location of the
   stack.  Advance YYPTR to a properly aligned location for the next
   stack.  */
# define YYSTACK_RELOCATE(Stack)					\
    do									\
      {									\
	YYSIZE_T yynewbytes;						\
	YYCOPY (&yyptr->Stack, Stack, yysize);				\
	Stack = &yyptr->Stack;						\
	yynewbytes = yystacksize * sizeof (*Stack) + YYSTACK_GAP_MAXIMUM; \
	yyptr += yynewbytes / sizeof (*yyptr);				\
      }									\
    while (YYID (0))

#endif

/* YYFINAL -- State number of the termination state.  */
#define YYFINAL  4
/* YYLAST -- Last index in YYTABLE.  */
#define YYLAST   2293

/* YYNTOKENS -- Number of terminals.  */
#define YYNTOKENS  76
/* YYNNTS -- Number of nonterminals.  */
#define YYNNTS  164
/* YYNRULES -- Number of rules.  */
#define YYNRULES  374
/* YYNRULES -- Number of states.  */
#define YYNSTATES  713

/* YYTRANSLATE(YYLEX) -- Bison symbol number corresponding to YYLEX.  */
#define YYUNDEFTOK  2
#define YYMAXUTOK   306

#define YYTRANSLATE(YYX)						\
  ((unsigned int) (YYX) <= YYMAXUTOK ? yytranslate[YYX] : YYUNDEFTOK)

/* YYTRANSLATE[YYLEX] -- Bison symbol number corresponding to YYLEX.  */
static const yytype_uint8 yytranslate[] =
{
       0,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    69,     2,     2,    64,    55,    56,     2,
      59,    60,    53,    49,    75,    50,    63,    54,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,    66,    62,
       2,    65,     2,    73,    74,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,    71,     2,    72,    52,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    67,    51,    68,    70,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     1,     2,     3,     4,
       5,     6,     7,     8,     9,    10,    11,    12,    13,    14,
      15,    16,    17,    18,    19,    20,    21,    22,    23,    24,
      25,    26,    27,    28,    29,    30,    31,    32,    33,    34,
      35,    36,    37,    38,    39,    40,    41,    42,    43,    44,
      45,    46,    47,    48,    57,    58,    61
};

#if YYDEBUG
/* YYPRHS[YYN] -- Index of the first RHS symbol of rule number YYN in
   YYRHS.  */
static const yytype_uint16 yyprhs[] =
{
       0,     0,     3,    19,    20,    24,    25,    29,    30,    34,
      35,    39,    40,    44,    45,    49,    50,    54,    55,    59,
      60,    64,    65,    69,    70,    74,    75,    79,    80,    84,
      85,    89,    92,    98,   102,   106,   109,   111,   115,   117,
     120,   123,   128,   129,   131,   132,   137,   138,   140,   142,
     144,   146,   149,   155,   159,   162,   168,   176,   180,   183,
     189,   193,   195,   198,   203,   207,   212,   216,   218,   221,
     223,   225,   228,   230,   234,   238,   242,   245,   248,   252,
     258,   264,   267,   268,   273,   274,   278,   279,   282,   283,
     288,   293,   298,   301,   307,   309,   311,   314,   315,   319,
     321,   325,   326,   327,   328,   337,   338,   344,   345,   348,
     349,   352,   353,   354,   362,   363,   369,   371,   375,   379,
     383,   387,   391,   395,   399,   403,   407,   411,   415,   419,
     423,   427,   431,   435,   439,   443,   447,   451,   453,   456,
     459,   462,   465,   468,   471,   474,   477,   481,   487,   494,
     496,   498,   502,   508,   514,   519,   526,   535,   537,   543,
     549,   555,   563,   565,   566,   570,   572,   577,   579,   584,
     586,   590,   592,   594,   596,   598,   600,   602,   604,   605,
     607,   609,   611,   613,   618,   623,   625,   627,   629,   632,
     634,   636,   638,   640,   642,   646,   648,   650,   652,   655,
     657,   659,   661,   663,   667,   669,   671,   673,   675,   677,
     679,   681,   683,   685,   689,   694,   699,   702,   706,   712,
     714,   716,   719,   723,   729,   733,   739,   743,   747,   753,
     762,   768,   777,   783,   784,   788,   789,   791,   795,   797,
     802,   805,   806,   810,   812,   816,   818,   822,   824,   828,
     830,   834,   836,   840,   844,   847,   852,   856,   862,   868,
     870,   874,   876,   879,   881,   885,   890,   892,   895,   898,
     900,   902,   906,   907,   910,   911,   913,   915,   917,   919,
     921,   923,   925,   927,   929,   930,   935,   937,   940,   943,
     946,   949,   952,   955,   957,   961,   963,   967,   969,   973,
     975,   979,   981,   985,   987,   989,   993,   997,   998,  1001,
    1002,  1004,  1005,  1007,  1008,  1010,  1011,  1013,  1014,  1016,
    1017,  1019,  1020,  1022,  1023,  1025,  1026,  1028,  1033,  1038,
    1044,  1051,  1056,  1061,  1063,  1065,  1067,  1069,  1071,  1073,
    1075,  1077,  1079,  1083,  1088,  1094,  1099,  1104,  1107,  1110,
    1115,  1119,  1123,  1129,  1133,  1138,  1142,  1148,  1150,  1151,
    1153,  1157,  1159,  1161,  1164,  1166,  1168,  1174,  1175,  1178,
    1180,  1184,  1186,  1190,  1192
};

/* YYRHS -- A `-1'-separated list of the rules' RHS.  */
static const yytype_int16 yyrhs[] =
{
      77,     0,    -1,    79,    99,   101,    89,    81,    91,    93,
      97,    95,    83,    85,    87,    78,   103,   188,    -1,    -1,
      25,   163,    62,    -1,    -1,    80,   108,   110,    -1,    -1,
      82,   108,   110,    -1,    -1,    84,   108,   110,    -1,    -1,
      86,   108,   110,    -1,    -1,    88,   108,   110,    -1,    -1,
      90,   108,   110,    -1,    -1,    92,   108,   110,    -1,    -1,
      94,   108,   110,    -1,    -1,    96,   108,   110,    -1,    -1,
      98,   108,   110,    -1,    -1,   100,   108,   110,    -1,    -1,
     102,   108,   110,    -1,    -1,   103,   104,    62,    -1,    21,
     105,    -1,    21,    59,   106,   212,    60,    -1,    21,    59,
      60,    -1,   107,   108,   110,    -1,   107,   110,    -1,   105,
      -1,   106,    62,   105,    -1,     3,    -1,   163,     3,    -1,
      63,     3,    -1,    25,    24,   109,    62,    -1,    -1,    24,
      -1,    -1,   111,   236,    64,    64,    -1,    -1,   113,    -1,
     180,    -1,   203,    -1,     1,    -1,    32,   115,    -1,    32,
      59,   189,   212,    60,    -1,    32,    59,    60,    -1,   114,
     116,    -1,   114,    59,   116,   212,    60,    -1,   114,    59,
     116,    62,   190,   212,    60,    -1,   114,    59,    60,    -1,
      31,   119,    -1,    31,    59,   191,   212,    60,    -1,    31,
      59,    60,    -1,     9,    -1,   207,   168,    -1,   207,   168,
      65,   208,    -1,   207,    65,   208,    -1,   207,   168,    65,
     208,    -1,   207,    65,   208,    -1,   116,    -1,   207,   168,
      -1,   207,    -1,   163,    -1,   118,   168,    -1,   148,    -1,
     148,     4,   148,    -1,   208,    65,   208,    -1,   208,     5,
     208,    -1,   148,    42,    -1,   148,    37,    -1,     7,   209,
      66,    -1,     7,   209,    65,   148,    66,    -1,     7,   209,
       5,   148,    66,    -1,    12,    66,    -1,    -1,    67,   123,
     205,    68,    -1,    -1,   121,   125,   205,    -1,    -1,   126,
     124,    -1,    -1,    35,   128,   205,    68,    -1,   208,    65,
      26,   148,    -1,   208,     5,    26,   148,    -1,    26,   148,
      -1,   216,    62,   216,    62,   216,    -1,   216,    -1,   129,
      -1,   130,   127,    -1,    -1,    16,   133,   131,    -1,   216,
      -1,   216,    62,   216,    -1,    -1,    -1,    -1,    20,   136,
     134,   137,   127,   138,   141,   142,    -1,    -1,    14,    20,
     140,   134,   127,    -1,    -1,   141,   139,    -1,    -1,    14,
     122,    -1,    -1,    -1,    30,   144,   134,   145,    35,   126,
      68,    -1,    -1,    28,   147,    35,   126,    68,    -1,   149,
      -1,   148,    47,   148,    -1,   148,    33,   148,    -1,   148,
      38,   148,    -1,   148,    46,   148,    -1,   148,    45,   148,
      -1,   148,    43,   148,    -1,   148,    39,   148,    -1,   148,
      40,   148,    -1,   148,    49,   148,    -1,   148,    50,   148,
      -1,   148,    51,   148,    -1,   148,    52,   148,    -1,   148,
      53,   148,    -1,   148,    54,   148,    -1,   148,    55,   148,
      -1,   148,    56,   148,    -1,   148,    34,   148,    -1,   148,
      44,   148,    -1,   148,    48,   148,    -1,   148,    36,   148,
      -1,   156,    -1,    53,   149,    -1,    56,   149,    -1,    49,
     149,    -1,    50,   149,    -1,    69,   149,    -1,    70,   149,
      -1,    52,   149,    -1,    36,   149,    -1,   156,    59,    60,
      -1,   156,    59,   209,   213,    60,    -1,   156,    59,   209,
      11,   213,    60,    -1,     3,    -1,   165,    -1,   156,    63,
     163,    -1,   156,    63,    59,   157,    60,    -1,   156,    63,
      59,    31,    60,    -1,   156,    71,   148,    72,    -1,   156,
      71,   214,    66,   214,    72,    -1,   156,    71,   214,    66,
     214,    66,   214,    72,    -1,   150,    -1,   171,    59,   148,
     213,    60,    -1,   172,   159,   152,   211,    68,    -1,   151,
      67,   152,   211,    68,    -1,    59,   157,    60,    67,   152,
     211,    68,    -1,   187,    -1,    -1,   148,    66,   155,    -1,
     148,    -1,    67,   152,   211,    68,    -1,   148,    -1,    67,
     152,   211,    68,    -1,   151,    -1,    59,   157,    60,    -1,
     148,    -1,   169,    -1,   168,    -1,    35,    -1,    67,    -1,
     163,    -1,   163,    -1,    -1,   160,    -1,    24,    -1,   164,
      -1,    73,    -1,    74,     3,    63,    24,    -1,    74,     3,
      63,    73,    -1,   163,    -1,   160,    -1,    11,    -1,    11,
     168,    -1,   177,    -1,   183,    -1,   175,    -1,   176,    -1,
     174,    -1,    59,   168,    60,    -1,   177,    -1,   183,    -1,
     175,    -1,    53,   169,    -1,   183,    -1,   175,    -1,   176,
      -1,   174,    -1,    59,   168,    60,    -1,   183,    -1,   175,
      -1,   175,    -1,   177,    -1,   183,    -1,   175,    -1,   176,
      -1,   174,    -1,   165,    -1,   165,    63,   163,    -1,    71,
     214,    72,   168,    -1,    71,    11,    72,   168,    -1,     8,
     170,    -1,     8,    36,   168,    -1,    23,    71,   168,    72,
     168,    -1,   178,    -1,   179,    -1,    53,   168,    -1,    36,
       8,   168,    -1,    29,   159,   192,   212,    68,    -1,    29,
     159,    68,    -1,    22,   159,   193,   212,    68,    -1,    22,
     159,    68,    -1,    17,   181,   184,    -1,   163,    59,   201,
      60,   185,    -1,    59,   201,    60,   163,    59,   201,    60,
     185,    -1,   222,    59,   217,    60,   232,    -1,    59,   237,
      60,   163,    59,   217,    60,   232,    -1,    17,    59,   201,
      60,   185,    -1,    -1,    67,   205,    68,    -1,    -1,   173,
      -1,    59,   201,    60,    -1,   183,    -1,   186,   159,   205,
      68,    -1,   186,     1,    -1,    -1,   188,   112,    62,    -1,
     115,    -1,   189,    62,   115,    -1,   117,    -1,   190,    62,
     117,    -1,   119,    -1,   191,    62,   119,    -1,   194,    -1,
     192,    62,   194,    -1,   197,    -1,   193,    62,   197,    -1,
     206,   168,   220,    -1,   196,   220,    -1,    59,   196,    60,
     220,    -1,    53,   196,   220,    -1,    59,    53,   196,    60,
     220,    -1,    53,    59,   196,    60,   220,    -1,    24,    -1,
      24,    63,   163,    -1,   195,    -1,   160,   198,    -1,   195,
      -1,    59,   195,    60,    -1,    59,   201,    60,   185,    -1,
     158,    -1,   163,   158,    -1,   163,   167,    -1,   167,    -1,
     199,    -1,   200,    75,   199,    -1,    -1,   200,   213,    -1,
      -1,   122,    -1,   113,    -1,   203,    -1,     1,    -1,   120,
      -1,   132,    -1,   143,    -1,   146,    -1,   135,    -1,    -1,
     166,    66,   204,   202,    -1,    15,    -1,     6,   162,    -1,
      10,   162,    -1,    18,   150,    -1,    13,   150,    -1,    19,
     160,    -1,    27,   215,    -1,   202,    -1,   205,    62,   202,
      -1,   160,    -1,   206,    75,   160,    -1,   161,    -1,   207,
      75,   161,    -1,   148,    -1,   208,    75,   148,    -1,   157,
      -1,   209,    75,   157,    -1,   153,    -1,   154,    -1,   210,
      75,   153,    -1,   210,    75,   154,    -1,    -1,   210,   213,
      -1,    -1,    62,    -1,    -1,    75,    -1,    -1,   148,    -1,
      -1,   208,    -1,    -1,   120,    -1,    -1,   237,    -1,    -1,
     238,    -1,    -1,   239,    -1,    -1,     3,    -1,    21,    24,
       3,    62,    -1,    32,   222,   224,    62,    -1,     9,   222,
      65,   235,    62,    -1,     9,   222,   224,    65,   235,    62,
      -1,    31,   223,   224,    62,    -1,    17,   182,   184,    62,
      -1,   164,    -1,   222,    -1,   226,    -1,   227,    -1,   228,
      -1,   226,    -1,   228,    -1,   164,    -1,    24,    -1,    71,
      72,   224,    -1,    71,     3,    72,   224,    -1,    23,    71,
     224,    72,   224,    -1,    29,    67,   218,    68,    -1,    22,
      67,   219,    68,    -1,    53,   224,    -1,     8,   225,    -1,
       8,    59,   227,    60,    -1,     8,    36,   224,    -1,    36,
       8,   224,    -1,    17,    59,   217,    60,   232,    -1,   163,
     224,   220,    -1,   163,    11,   224,   220,    -1,   163,   224,
     220,    -1,   163,    59,   217,    60,   232,    -1,   224,    -1,
      -1,   233,    -1,    59,   217,    60,    -1,   224,    -1,     3,
      -1,    50,     3,    -1,   163,    -1,   234,    -1,    59,   234,
      49,   234,    60,    -1,    -1,   236,   221,    -1,   229,    -1,
     237,    75,   229,    -1,   230,    -1,   238,    62,   230,    -1,
     231,    -1,   239,    62,   231,    -1
};

/* YYRLINE[YYN] -- source line where rule number YYN was defined.  */
static const yytype_uint16 yyrline[] =
{
       0,   124,   124,   144,   150,   161,   161,   178,   178,   196,
     196,   214,   214,   232,   232,   250,   250,   268,   268,   285,
     285,   303,   303,   321,   321,   339,   339,   362,   362,   378,
     379,   382,   383,   384,   387,   424,   435,   436,   439,   446,
     453,   462,   476,   477,   484,   484,   497,   501,   502,   506,
     511,   517,   521,   525,   529,   535,   541,   547,   552,   556,
     560,   566,   572,   576,   580,   586,   590,   596,   597,   601,
     607,   616,   622,   640,   645,   657,   673,   679,   687,   707,
     725,   734,   753,   752,   767,   766,   798,   801,   808,   807,
     818,   824,   831,   838,   849,   855,   858,   866,   865,   876,
     882,   894,   898,   903,   893,   924,   923,   936,   939,   945,
     948,   960,   964,   959,   982,   981,   997,   998,  1002,  1006,
    1010,  1014,  1018,  1022,  1026,  1030,  1034,  1038,  1042,  1046,
    1050,  1054,  1058,  1062,  1066,  1070,  1075,  1081,  1082,  1086,
    1097,  1101,  1105,  1109,  1114,  1118,  1128,  1132,  1137,  1145,
    1149,  1150,  1161,  1165,  1169,  1173,  1177,  1185,  1186,  1192,
    1199,  1205,  1212,  1215,  1222,  1228,  1245,  1252,  1253,  1260,
    1261,  1280,  1281,  1284,  1287,  1291,  1302,  1311,  1317,  1320,
    1323,  1330,  1331,  1337,  1350,  1365,  1373,  1385,  1390,  1396,
    1397,  1398,  1399,  1400,  1401,  1407,  1408,  1409,  1410,  1416,
    1417,  1418,  1419,  1420,  1426,  1427,  1430,  1433,  1434,  1435,
    1436,  1437,  1440,  1441,  1454,  1458,  1463,  1468,  1473,  1477,
    1478,  1481,  1487,  1494,  1500,  1507,  1513,  1524,  1540,  1569,
    1607,  1632,  1650,  1659,  1662,  1670,  1674,  1678,  1685,  1691,
    1696,  1708,  1711,  1723,  1724,  1730,  1731,  1737,  1741,  1747,
    1748,  1754,  1758,  1764,  1787,  1792,  1798,  1804,  1811,  1820,
    1829,  1844,  1850,  1855,  1859,  1866,  1879,  1880,  1886,  1892,
    1895,  1899,  1905,  1908,  1917,  1920,  1921,  1925,  1926,  1932,
    1933,  1934,  1935,  1936,  1938,  1937,  1952,  1958,  1962,  1966,
    1970,  1974,  1979,  1998,  2004,  2012,  2016,  2022,  2026,  2032,
    2036,  2042,  2046,  2055,  2059,  2063,  2067,  2073,  2076,  2084,
    2085,  2087,  2088,  2091,  2094,  2097,  2100,  2103,  2106,  2109,
    2112,  2115,  2118,  2121,  2124,  2127,  2130,  2136,  2140,  2144,
    2148,  2152,  2156,  2176,  2183,  2194,  2195,  2196,  2199,  2200,
    2203,  2207,  2217,  2221,  2225,  2229,  2233,  2237,  2241,  2247,
    2253,  2261,  2269,  2275,  2282,  2298,  2320,  2324,  2330,  2333,
    2336,  2340,  2350,  2354,  2373,  2381,  2382,  2394,  2395,  2398,
    2402,  2408,  2412,  2418,  2422
};
#endif

#if YYDEBUG || YYERROR_VERBOSE || YYTOKEN_TABLE
/* YYTNAME[SYMBOL-NUM] -- String name of the symbol SYMBOL-NUM.
   First, the terminals, then, starting at YYNTOKENS, nonterminals.  */
const char *yytname[] =
{
  "$end", "error", "$undefined", "LLITERAL", "LASOP", "LCOLAS", "LBREAK",
  "LCASE", "LCHAN", "LCONST", "LCONTINUE", "LDDD", "LDEFAULT", "LDEFER",
  "LELSE", "LFALL", "LFOR", "LFUNC", "LGO", "LGOTO", "LIF", "LIMPORT",
  "LINTERFACE", "LMAP", "LNAME", "LPACKAGE", "LRANGE", "LRETURN",
  "LSELECT", "LSTRUCT", "LSWITCH", "LTYPE", "LVAR", "LANDAND", "LANDNOT",
  "LBODY", "LCOMM", "LDEC", "LEQ", "LGE", "LGT", "LIGNORE", "LINC", "LLE",
  "LLSH", "LLT", "LNE", "LOROR", "LRSH", "'+'", "'-'", "'|'", "'^'", "'*'",
  "'/'", "'%'", "'&'", "NotPackage", "NotParen", "'('", "')'",
  "PreferToRightParen", "';'", "'.'", "'$'", "'='", "':'", "'{'", "'}'",
  "'!'", "'~'", "'['", "']'", "'?'", "'@'", "','", "$accept", "file",
  "package", "loadcore", "@1", "loadchannels", "@2", "loadseq", "@3",
  "loaddefers", "@4", "loadruntime", "@5", "loadsched", "@6", "loadhash",
  "@7", "loadprintf", "@8", "loadifacestuff", "@9", "loadstrings", "@10",
  "loadmaps", "@11", "loadwb", "@12", "imports", "import", "import_stmt",
  "import_stmt_list", "import_here", "import_package", "import_safety",
  "import_there", "@13", "xdcl", "common_dcl", "lconst", "vardcl",
  "constdcl", "constdcl1", "typedclname", "typedcl", "simple_stmt", "case",
  "compound_stmt", "@14", "caseblock", "@15", "caseblock_list",
  "loop_body", "@16", "range_stmt", "for_header", "for_body", "for_stmt",
  "@17", "if_header", "if_stmt", "@18", "@19", "@20", "elseif", "@21",
  "elseif_list", "else", "switch_stmt", "@22", "@23", "select_stmt", "@24",
  "expr", "uexpr", "pseudocall", "pexpr_no_paren", "start_complit",
  "keyval", "bare_complitexpr", "complitexpr", "pexpr", "expr_or_type",
  "name_or_type", "lbrace", "new_name", "dcl_name", "onew_name", "sym",
  "hidden_importsym", "name", "labelname", "dotdotdot", "ntype",
  "non_expr_type", "non_recvchantype", "convtype", "comptype",
  "fnret_type", "dotname", "othertype", "ptrtype", "recvchantype",
  "structtype", "interfacetype", "xfndcl", "fndcl", "hidden_fndcl",
  "fntype", "fnbody", "fnres", "fnlitdcl", "fnliteral", "xdcl_list",
  "vardcl_list", "constdcl_list", "typedcl_list", "structdcl_list",
  "interfacedcl_list", "structdcl", "packname", "embed", "interfacedcl",
  "indcl", "arg_type", "arg_type_list", "oarg_type_list_ocomma", "stmt",
  "non_dcl_stmt", "@25", "stmt_list", "new_name_list", "dcl_name_list",
  "expr_list", "expr_or_type_list", "keyval_list", "braced_keyval_list",
  "osemi", "ocomma", "oexpr", "oexpr_list", "osimple_stmt",
  "ohidden_funarg_list", "ohidden_structdcl_list",
  "ohidden_interfacedcl_list", "oliteral", "hidden_import",
  "hidden_pkg_importsym", "hidden_pkgtype", "hidden_type",
  "hidden_type_non_recv_chan", "hidden_type_misc", "hidden_type_recv_chan",
  "hidden_type_func", "hidden_funarg", "hidden_structdcl",
  "hidden_interfacedcl", "ohidden_funres", "hidden_funres",
  "hidden_literal", "hidden_constant", "hidden_import_list",
  "hidden_funarg_list", "hidden_structdcl_list",
  "hidden_interfacedcl_list", 0
};
#endif

# ifdef YYPRINT
/* YYTOKNUM[YYLEX-NUM] -- Internal token number corresponding to
   token YYLEX-NUM.  */
static const yytype_uint16 yytoknum[] =
{
       0,   256,   257,   258,   259,   260,   261,   262,   263,   264,
     265,   266,   267,   268,   269,   270,   271,   272,   273,   274,
     275,   276,   277,   278,   279,   280,   281,   282,   283,   284,
     285,   286,   287,   288,   289,   290,   291,   292,   293,   294,
     295,   296,   297,   298,   299,   300,   301,   302,   303,    43,
      45,   124,    94,    42,    47,    37,    38,   304,   305,    40,
      41,   306,    59,    46,    36,    61,    58,   123,   125,    33,
     126,    91,    93,    63,    64,    44
};
# endif

/* YYR1[YYN] -- Symbol number of symbol that rule YYN derives.  */
static const yytype_uint8 yyr1[] =
{
       0,    76,    77,    78,    78,    80,    79,    82,    81,    84,
      83,    86,    85,    88,    87,    90,    89,    92,    91,    94,
      93,    96,    95,    98,    97,   100,    99,   102,   101,   103,
     103,   104,   104,   104,   105,   105,   106,   106,   107,   107,
     107,   108,   109,   109,   111,   110,   112,   112,   112,   112,
     112,   113,   113,   113,   113,   113,   113,   113,   113,   113,
     113,   114,   115,   115,   115,   116,   116,   117,   117,   117,
     118,   119,   120,   120,   120,   120,   120,   120,   121,   121,
     121,   121,   123,   122,   125,   124,   126,   126,   128,   127,
     129,   129,   129,   130,   130,   130,   131,   133,   132,   134,
     134,   136,   137,   138,   135,   140,   139,   141,   141,   142,
     142,   144,   145,   143,   147,   146,   148,   148,   148,   148,
     148,   148,   148,   148,   148,   148,   148,   148,   148,   148,
     148,   148,   148,   148,   148,   148,   148,   149,   149,   149,
     149,   149,   149,   149,   149,   149,   150,   150,   150,   151,
     151,   151,   151,   151,   151,   151,   151,   151,   151,   151,
     151,   151,   151,   152,   153,   154,   154,   155,   155,   156,
     156,   157,   157,   158,   159,   159,   160,   161,   162,   162,
     163,   163,   163,   164,   164,   165,   166,   167,   167,   168,
     168,   168,   168,   168,   168,   169,   169,   169,   169,   170,
     170,   170,   170,   170,   171,   171,   172,   173,   173,   173,
     173,   173,   174,   174,   175,   175,   175,   175,   175,   175,
     175,   176,   177,   178,   178,   179,   179,   180,   181,   181,
     182,   182,   183,   184,   184,   185,   185,   185,   186,   187,
     187,   188,   188,   189,   189,   190,   190,   191,   191,   192,
     192,   193,   193,   194,   194,   194,   194,   194,   194,   195,
     195,   196,   197,   197,   197,   198,   199,   199,   199,   199,
     200,   200,   201,   201,   202,   202,   202,   202,   202,   203,
     203,   203,   203,   203,   204,   203,   203,   203,   203,   203,
     203,   203,   203,   205,   205,   206,   206,   207,   207,   208,
     208,   209,   209,   210,   210,   210,   210,   211,   211,   212,
     212,   213,   213,   214,   214,   215,   215,   216,   216,   217,
     217,   218,   218,   219,   219,   220,   220,   221,   221,   221,
     221,   221,   221,   222,   223,   224,   224,   224,   225,   225,
     226,   226,   226,   226,   226,   226,   226,   226,   226,   226,
     226,   227,   228,   229,   229,   230,   231,   231,   232,   232,
     233,   233,   234,   234,   234,   235,   235,   236,   236,   237,
     237,   238,   238,   239,   239
};

/* YYR2[YYN] -- Number of symbols composing right hand side of rule YYN.  */
static const yytype_uint8 yyr2[] =
{
       0,     2,    15,     0,     3,     0,     3,     0,     3,     0,
       3,     0,     3,     0,     3,     0,     3,     0,     3,     0,
       3,     0,     3,     0,     3,     0,     3,     0,     3,     0,
       3,     2,     5,     3,     3,     2,     1,     3,     1,     2,
       2,     4,     0,     1,     0,     4,     0,     1,     1,     1,
       1,     2,     5,     3,     2,     5,     7,     3,     2,     5,
       3,     1,     2,     4,     3,     4,     3,     1,     2,     1,
       1,     2,     1,     3,     3,     3,     2,     2,     3,     5,
       5,     2,     0,     4,     0,     3,     0,     2,     0,     4,
       4,     4,     2,     5,     1,     1,     2,     0,     3,     1,
       3,     0,     0,     0,     8,     0,     5,     0,     2,     0,
       2,     0,     0,     7,     0,     5,     1,     3,     3,     3,
       3,     3,     3,     3,     3,     3,     3,     3,     3,     3,
       3,     3,     3,     3,     3,     3,     3,     1,     2,     2,
       2,     2,     2,     2,     2,     2,     3,     5,     6,     1,
       1,     3,     5,     5,     4,     6,     8,     1,     5,     5,
       5,     7,     1,     0,     3,     1,     4,     1,     4,     1,
       3,     1,     1,     1,     1,     1,     1,     1,     0,     1,
       1,     1,     1,     4,     4,     1,     1,     1,     2,     1,
       1,     1,     1,     1,     3,     1,     1,     1,     2,     1,
       1,     1,     1,     3,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     3,     4,     4,     2,     3,     5,     1,
       1,     2,     3,     5,     3,     5,     3,     3,     5,     8,
       5,     8,     5,     0,     3,     0,     1,     3,     1,     4,
       2,     0,     3,     1,     3,     1,     3,     1,     3,     1,
       3,     1,     3,     3,     2,     4,     3,     5,     5,     1,
       3,     1,     2,     1,     3,     4,     1,     2,     2,     1,
       1,     3,     0,     2,     0,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     0,     4,     1,     2,     2,     2,
       2,     2,     2,     1,     3,     1,     3,     1,     3,     1,
       3,     1,     3,     1,     1,     3,     3,     0,     2,     0,
       1,     0,     1,     0,     1,     0,     1,     0,     1,     0,
       1,     0,     1,     0,     1,     0,     1,     4,     4,     5,
       6,     4,     4,     1,     1,     1,     1,     1,     1,     1,
       1,     1,     3,     4,     5,     4,     4,     2,     2,     4,
       3,     3,     5,     3,     4,     3,     5,     1,     0,     1,
       3,     1,     1,     2,     1,     1,     5,     0,     2,     1,
       3,     1,     3,     1,     3
};

/* YYDEFACT[STATE-NAME] -- Default rule to reduce with in state
   STATE-NUM when YYTABLE doesn't specify something else to do.  Zero
   means the default is an error.  */
static const yytype_uint16 yydefact[] =
{
       5,     0,    25,     0,     1,    27,     0,     0,    44,    15,
       0,    44,    42,     6,   367,     7,     0,    44,    26,    43,
       0,     0,    17,     0,    44,    28,    41,     0,     0,     0,
       0,     0,     0,   368,    19,     0,    44,    16,     0,   333,
       0,     0,   233,     0,     0,   334,     0,     0,    45,    23,
       0,    44,     8,     0,     0,     0,     0,     0,   341,     0,
       0,     0,     0,     0,   340,     0,   335,   336,   337,   180,
     182,     0,   181,   369,     0,     0,     0,   319,     0,     0,
       0,    21,     0,    44,    18,     0,     0,     0,   348,   338,
     339,   319,   323,     0,   321,     0,   347,   362,     0,     0,
     364,   365,     0,     0,     0,     0,     0,   325,     0,     0,
     278,   149,   178,     0,    61,   178,     0,   286,    97,     0,
       0,     0,   101,     0,     0,   315,   114,     0,   111,     0,
       0,     0,     0,     0,     0,     0,     0,     0,    82,     0,
       0,   313,   276,     0,   279,   275,   280,   283,   281,   282,
      72,   116,   157,   169,   137,   186,   185,   150,     0,     0,
       0,   206,   219,   220,   238,     0,   162,   293,   277,     0,
       0,   332,     0,   320,   327,   331,   328,     9,     0,    44,
      20,   183,   184,   350,     0,     0,   341,     0,   340,     0,
     357,   373,   324,     0,     0,     0,   371,   322,   351,   363,
       0,   329,     0,   342,     0,   325,   326,   353,     0,   370,
     179,   287,   176,     0,     0,     0,   185,   212,   216,   202,
     200,   201,   199,   288,   157,     0,   317,   272,   157,   291,
     317,   174,   175,     0,     0,   299,   316,   292,     0,     0,
     317,     0,     0,    58,    70,     0,    51,   297,   177,     0,
     145,   140,   141,   144,   138,   139,     0,     0,   171,     0,
     172,   197,   195,   196,     0,   142,   143,     0,   314,     0,
       0,    54,     0,     0,     0,     0,     0,    77,     0,     0,
       0,    76,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,   163,     0,     0,   313,
     284,     0,   163,   240,     0,     0,   234,     0,     0,     0,
     358,    11,     0,    44,    24,   349,   358,   319,   346,     0,
       0,   325,   345,     0,     0,   343,   330,   354,   319,     0,
       0,   217,   193,   191,   192,   189,   190,   221,     0,     0,
       0,   318,    95,     0,    98,     0,    94,   187,   266,   185,
     269,   173,   270,   311,     0,   102,    99,   180,     0,   226,
       0,   309,   263,   251,     0,    86,     0,     0,   224,   295,
     309,   249,   261,   325,     0,   112,    60,   247,   309,    71,
      53,   243,   309,     0,     0,    62,     0,   198,   170,     0,
       0,     0,    57,   309,     0,     0,    73,   118,   133,   136,
     119,   123,   124,   122,   134,   121,   120,   117,   135,   125,
     126,   127,   128,   129,   130,   131,   132,   307,   146,   301,
     311,     0,   151,   314,     0,     0,   311,   307,     0,   294,
      75,    74,   300,   319,   361,   230,   359,    13,     0,    44,
      22,   352,     0,   374,   344,   355,   372,     0,     0,     0,
       0,   203,   213,    92,    88,    96,     0,     0,   317,   188,
     267,   268,   312,   273,   235,     0,   317,     0,   259,     0,
     272,   262,   310,     0,     0,     0,     0,   325,     0,     0,
     310,     0,   254,     0,   325,     0,   310,     0,   310,     0,
      64,   298,     0,     0,     0,   222,   193,   191,   192,   190,
     163,    83,   215,   214,   310,     0,    66,     0,   163,   165,
     303,   304,   311,     0,   311,   312,     0,     0,     0,   154,
     313,   285,   312,     0,     0,   239,     0,     3,     0,    44,
      10,   358,   366,   358,   194,     0,     0,     0,     0,   271,
     272,   236,   211,   209,   210,   207,   208,   232,   103,   100,
     260,   264,     0,   252,   225,   218,     0,     0,   115,    84,
      87,     0,   256,     0,   325,   250,   223,   296,   253,    86,
     248,    59,   244,    52,    63,     0,   307,    67,   245,   309,
      69,    55,    65,   307,     0,   312,   308,   160,     0,   302,
     147,   153,   152,     0,   158,   159,   360,     0,    29,    44,
      12,   356,   231,     0,    91,    90,   317,     0,   107,   235,
       0,    81,     0,   325,   325,   255,     0,   194,     0,   310,
       0,    68,     0,   163,   167,   164,   305,   306,   148,   313,
     155,     0,   241,    14,    89,    93,   237,   109,   265,     0,
       0,    78,     0,    85,   258,   257,   113,   161,   246,    56,
     166,   307,     0,     4,     0,     0,     0,     0,   108,   104,
       0,     0,     0,   156,    38,     0,     0,    31,    44,     0,
      30,    50,     0,     0,    47,    48,    49,   105,   110,    80,
      79,   168,    33,    36,   309,    40,    44,    35,    39,   272,
       0,   233,   242,   317,   310,     0,    34,     0,   272,   227,
       0,    37,    32,   235,     0,   106,   185,   235,   272,   228,
       0,   235,   229
};

/* YYDEFGOTO[NTERM-NUM].  */
static const yytype_int16 yydefgoto[] =
{
      -1,     1,   598,     2,     3,    22,    23,   311,   312,   437,
     438,   527,   528,    15,    16,    34,    35,    49,    50,   177,
     178,    81,    82,     5,     6,     9,    10,   632,   655,   667,
     684,   668,     8,    20,    13,    14,   673,   142,   143,   246,
     577,   578,   242,   243,   144,   559,   145,   264,   560,   612,
     475,   455,   535,   342,   343,   344,   146,   226,   355,   147,
     230,   465,   608,   658,   693,   637,   659,   148,   240,   485,
     149,   238,   150,   151,   152,   153,   417,   510,   511,   625,
     154,   419,   348,   233,   155,   247,   211,   216,    72,   157,
     158,   350,   351,   260,   218,   159,   160,   541,   332,   161,
     334,   335,   162,   163,   675,   691,    42,   164,    76,   547,
     165,   166,   656,   382,   579,   378,   370,   361,   371,   372,
     373,   363,   471,   352,   353,   354,   167,   168,   425,   169,
     374,   249,   170,   420,   512,   513,   473,   463,   269,   237,
     356,   172,   195,   189,   207,    33,    40,    46,   434,    88,
      66,    67,    68,    73,   196,   191,   435,   436,   101,   102,
      21,   173,   197,   192
};

/* YYPACT[STATE-NUM] -- Index in YYTABLE of the portion describing
   STATE-NUM.  */
#define YYPACT_NINF -574
static const yytype_int16 yypact[] =
{
    -574,    78,  -574,    61,  -574,  -574,    61,    65,  -574,  -574,
      61,  -574,    86,  -574,  -574,  -574,    61,  -574,  -574,  -574,
      40,   426,  -574,    61,  -574,  -574,  -574,    29,   113,   105,
      29,    29,    37,  -574,  -574,    61,  -574,  -574,   138,  -574,
     956,   253,   100,   139,   181,  -574,  1347,  1347,  -574,  -574,
      61,  -574,  -574,   146,  1936,   158,   152,   153,  -574,   171,
     251,  1347,    49,    43,  -574,   203,  -574,  -574,  -574,  -574,
    -574,  1962,  -574,  -574,   115,  1240,   210,   253,   214,   233,
     238,  -574,    61,  -574,  -574,    39,  1347,   267,  -574,  -574,
    -574,   253,  1989,  1347,   253,  1347,  -574,  -574,   313,    57,
    -574,  -574,   255,   260,  1347,    49,  1347,   334,   253,   253,
    -574,  -574,   253,  1842,  -574,   253,   140,  -574,  -574,   290,
     140,   253,  -574,    31,   282,  1654,  -574,    31,  -574,   240,
     300,  1654,  1654,  1654,  1654,  1654,  1654,  1697,  -574,  1654,
    1654,   847,  -574,   390,  -574,  -574,  -574,  -574,  -574,  -574,
    1282,  -574,  -574,   287,     5,  -574,   301,  -574,   305,   321,
      31,   325,  -574,  -574,   326,   187,  -574,  -574,  -574,   132,
      77,  -574,   330,   324,  -574,  -574,  -574,  -574,    61,  -574,
    -574,  -574,  -574,  -574,   336,   345,   347,   350,   354,   355,
    -574,  -574,   368,   359,  1347,   365,  -574,   375,  -574,  -574,
     392,  -574,  1347,  -574,   394,   334,  -574,  -574,   403,  -574,
    -574,  -574,  -574,  1869,  1869,  1869,  -574,   404,  -574,  -574,
    -574,  -574,  -574,  -574,   216,     5,   996,  1816,   250,  -574,
    1654,  -574,  -574,   414,  1869,  2189,   399,  -574,   434,   348,
    1654,   283,  1869,  -574,  -574,   386,  -574,  -574,  -574,   869,
    -574,  -574,  -574,  -574,  -574,  -574,  1740,  1697,  2189,   415,
    -574,    20,  -574,    92,  1240,  -574,  -574,   411,  2189,   419,
     420,  -574,  1783,  1654,  1654,  1654,  1654,  -574,  1654,  1654,
    1654,  -574,  1654,  1654,  1654,  1654,  1654,  1654,  1654,  1654,
    1654,  1654,  1654,  1654,  1654,  1654,  -574,  1336,   462,  1654,
    -574,  1654,  -574,  -574,  1240,  1155,  -574,  1654,  1654,  1654,
    2015,  -574,    61,  -574,  -574,  -574,  2015,   253,  -574,  1989,
    1347,   334,  -574,   253,    57,  -574,  -574,  -574,   253,   468,
    1869,  -574,  -574,  -574,  -574,  -574,  -574,  -574,   436,   253,
    1654,  -574,  -574,   463,  -574,    79,   438,  1869,  -574,  1816,
    -574,  -574,  -574,   429,   441,  -574,   452,   103,   492,  -574,
     458,   457,  -574,  -574,   448,  -574,    26,   125,  -574,  -574,
     461,  -574,  -574,   334,  1808,  -574,  -574,  -574,   470,  -574,
    -574,  -574,   471,  1654,   253,   478,  1895,  -574,   460,   225,
    1869,  1869,  -574,   483,  1654,   481,  2189,  2027,  -574,  2213,
    1072,  1072,  1072,  1072,  -574,  1072,  1072,  2237,  -574,   507,
     507,   507,   507,  -574,  -574,  -574,  -574,  1391,  -574,  -574,
      38,  1446,  -574,  2087,   464,  1155,  2054,  1391,   243,  -574,
     399,   399,  2189,   253,  -574,  -574,  -574,  -574,    61,  -574,
    -574,  -574,   480,  -574,  -574,  -574,  -574,   490,   493,  1869,
     494,  -574,  -574,  2189,  -574,  -574,  1501,  1556,  1654,  -574,
    -574,  -574,  1816,  -574,  1903,   463,  1654,   253,   489,   496,
    1816,  -574,   465,   497,  1869,   102,   492,   334,   492,   504,
     157,   498,  -574,   253,   334,   533,   253,   512,   253,   515,
     399,  -574,  1654,  1928,  1869,  -574,   256,   263,   316,   327,
    -574,  -574,  -574,  -574,   253,   516,   399,  1654,  -574,  2117,
    -574,  -574,   502,   510,   505,  1697,   519,   522,   524,  -574,
    1654,  -574,  -574,   527,   521,  -574,   532,   568,    61,  -574,
    -574,  2015,  -574,  2015,  -574,  1240,  1654,  1654,   534,  -574,
    1816,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,
    -574,  -574,   538,  -574,  -574,  -574,  1697,   536,  -574,  -574,
    -574,   544,  -574,   546,   334,  -574,  -574,  -574,  -574,  -574,
    -574,  -574,  -574,  -574,   399,   548,  1391,  -574,  -574,   547,
    1783,  -574,   399,  1391,  1599,  1391,  -574,  -574,   550,  -574,
    -574,  -574,  -574,   111,  -574,  -574,  -574,   253,  -574,  -574,
    -574,  -574,  -574,   271,  2189,  2189,  1654,   551,  -574,  1903,
      70,  -574,  1155,   334,   334,  -574,   167,   344,   545,   253,
     552,   481,   549,  -574,  2189,  -574,  -574,  -574,  -574,  1654,
    -574,   557,   599,  -574,  -574,  -574,  -574,   607,  -574,  1654,
    1654,  -574,  1697,   560,  -574,  -574,  -574,  -574,  -574,  -574,
    -574,  1391,   554,  -574,   123,   561,  1081,    27,  -574,  -574,
    2141,  2165,   556,  -574,  -574,   246,   625,  -574,    61,   626,
    -574,  -574,   475,   569,  -574,  -574,  -574,  -574,  -574,  -574,
    -574,  -574,  -574,  -574,   570,  -574,  -574,  -574,  -574,  1816,
     571,   100,  -574,  1654,   268,   575,  -574,   577,  1816,  -574,
     463,  -574,  -574,  1903,   578,  -574,   580,  1903,  1816,  -574,
     581,  1903,  -574
};

/* YYPGOTO[NTERM-NUM].  */
static const yytype_int16 yypgoto[] =
{
    -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,
    -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,
    -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -492,
    -574,  -574,    -2,  -574,   -11,  -574,  -574,   -16,  -574,  -227,
    -112,    24,  -574,  -211,  -214,  -574,   -12,  -574,  -574,  -574,
      80,  -431,  -574,  -574,  -574,  -574,  -574,  -574,  -239,  -574,
    -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,
    -574,  -574,   432,   -15,   -58,  -574,  -279,    62,    63,  -574,
     -51,  -127,   297,   -70,   -83,   266,   537,   -38,   721,   295,
    -574,   304,   -22,   397,  -574,  -574,  -574,  -574,  -106,    -9,
    -104,  -117,  -574,  -574,  -574,  -574,  -574,   121,   -36,  -573,
    -574,  -574,  -574,  -574,  -574,  -574,  -574,  -574,   178,  -198,
    -323,   190,  -574,   197,  -574,  -443,  -264,     7,  -574,  -262,
    -574,  -138,    21,   108,  -574,  -368,  -282,  -375,  -284,  -574,
    -215,   -72,  -574,  -574,  -188,  -574,   389,  -574,   759,  -574,
     613,   584,   614,   563,   351,   357,  -294,  -574,   -71,   573,
    -574,   632,  -574,  -574
};

/* YYTABLE[YYPACT[STATE-NUM]].  What to do in state STATE-NUM.  If
   positive, shift that token.  If negative, reduce the rule which
   number is the opposite.  If zero, do what YYDEFACT says.
   If YYTABLE_NINF, syntax error.  */
#define YYTABLE_NINF -300
static const yytype_int16 yytable[] =
{
      18,   375,   389,    71,    11,   272,    25,   219,    17,   221,
     259,   346,   341,    37,    24,   424,   341,   327,   381,   185,
     262,    36,   441,   427,   100,    52,   341,   552,   200,   210,
     377,   271,   210,    51,   548,   362,   638,   156,   229,    71,
      84,   429,   428,   477,   479,   516,   103,   677,    83,   514,
     468,   523,    97,    71,   187,  -206,   194,   239,   224,   524,
      97,   100,   228,   181,   297,   225,   231,   100,   298,   225,
     208,    71,   180,    69,   212,   639,   299,   212,     4,  -205,
     179,    69,   307,   212,   456,   476,     7,  -206,   481,    12,
     302,   244,   248,  -238,   138,   304,   487,   607,   232,    98,
     489,    48,    26,    38,   220,   248,  -259,    98,    99,   556,
      19,   505,   182,   515,   557,   104,   250,   251,   252,   253,
     254,   255,    70,    38,   265,   266,   664,  -238,   261,    44,
      70,    38,   272,   445,   709,   640,   641,   586,   712,   588,
     262,    53,   308,   111,   457,   642,   236,    69,   113,   468,
     360,  -204,   309,   561,   309,   563,   369,   119,   393,  -238,
     469,   521,   123,   124,    69,  -259,   467,    75,   314,   127,
     558,  -259,    41,   683,   556,   108,   313,   629,   478,   557,
     262,   357,   665,   630,    78,   482,   666,    38,   303,   349,
     109,   331,   337,   338,   305,   212,    70,    38,    77,   137,
     306,   212,   701,   244,   333,   333,   333,   248,   618,    85,
     366,   141,   364,    70,    38,   622,   367,    91,   333,    92,
     379,   576,   231,  -290,    93,   333,   156,   385,  -290,   583,
      70,    38,   248,   333,   222,   646,   593,   601,    94,   602,
     333,   250,   254,   538,   341,   442,   697,   345,   261,   664,
     395,   549,   341,   447,   232,   704,   448,  -289,   263,    95,
     422,   572,  -289,   333,    69,   710,   156,   156,   105,   705,
      69,   664,   171,   603,   362,   570,   174,    69,  -290,    71,
     496,   187,   498,   662,  -290,   194,   100,   305,   261,   562,
      71,  -202,    69,   501,   518,   175,   568,   620,  -200,   241,
     176,   452,   440,    60,   262,   305,   682,    69,   450,   666,
     439,   525,  -289,    70,    38,  -202,   199,   201,  -289,    70,
      38,   333,  -200,  -202,    69,   459,    70,    38,   430,   431,
    -200,   666,   202,   305,   336,   336,   336,   206,   333,   634,
     333,    70,    38,   376,   651,   652,   248,   545,   336,   227,
     643,  -201,   484,   234,   296,   336,    70,    38,   542,   245,
     544,   526,  -199,   336,   495,   333,   580,  -176,   502,   503,
     336,   300,   357,    70,    38,  -201,   615,   497,   263,  -203,
     301,   333,   333,  -201,  -205,  -204,  -199,   156,   589,   360,
     310,   635,   341,   336,  -199,    71,   315,   369,   262,   109,
     567,   366,   695,  -203,   490,   316,  -180,   367,   217,   317,
      69,  -203,   261,  -181,    69,   506,   368,    43,   263,    45,
      47,    70,    38,   318,   349,   644,   645,   495,   530,   550,
     319,   320,   349,   322,   212,    27,   529,   323,   357,   262,
     333,   324,   212,    28,    69,   212,   380,    29,   244,   270,
     248,   336,   555,   333,   700,   543,   326,    30,    31,    70,
      38,   333,   328,    70,    38,   333,   248,   339,   336,   365,
     336,   331,   575,   358,   309,   388,   449,   430,   431,   341,
     392,   580,   359,   390,   333,   333,    69,    70,    38,   357,
      32,   391,   545,    70,    38,   336,   451,   156,   454,    69,
     458,   464,   349,   542,   462,   544,   261,   499,   217,   217,
     217,   336,   336,   574,   466,   589,   468,   470,   600,   472,
     474,   421,   217,   480,   358,   262,   599,   500,   582,   217,
     520,   333,   486,   488,   689,    70,    38,   217,    70,    38,
     531,   275,   263,   492,   217,   504,   507,   261,    70,    38,
     532,   283,   467,   533,   534,   287,   551,   235,   621,   631,
     292,   293,   294,   295,   564,   554,   566,   217,   569,   258,
     336,   333,   571,   268,   156,   573,   581,   585,   587,   590,
     522,   248,   591,   336,   592,   546,   545,   594,   633,   595,
     545,   336,   596,   597,   545,   336,   606,   542,   609,   544,
     543,   542,   611,   544,   613,   542,   614,   544,   617,   619,
     628,   636,   649,   647,   336,   336,   669,   650,   156,   653,
     654,   657,   305,   670,   681,   217,   663,   669,   685,   688,
     698,   692,   694,   261,   690,   702,   263,   703,   707,   708,
     674,   711,   217,   648,   217,   678,   460,   626,   627,   616,
     491,   349,   223,   461,   387,   699,   669,   687,   565,   539,
     349,   336,   553,   676,   610,   706,   686,    89,    90,   217,
     349,   184,   209,    74,   446,   696,   443,   263,   204,     0,
     333,   217,     0,     0,     0,   217,   217,     0,     0,   333,
       0,     0,     0,     0,   543,     0,     0,     0,   543,   333,
       0,   336,   543,     0,     0,   396,   397,   398,   399,     0,
     400,   401,   402,     0,   403,   404,   405,   406,   407,   408,
     409,   410,   411,   412,   413,   414,   415,   416,     0,   258,
     546,   423,     0,   426,     0,     0,     0,     0,     0,   235,
     235,   432,     0,     0,   217,     0,     0,     0,    39,    39,
       0,    39,    39,     0,     0,     0,     0,   217,     0,   217,
       0,    64,     0,   263,     0,   217,     0,    64,    64,   217,
       0,     0,   453,     0,     0,    64,     0,     0,     0,     0,
       0,     0,    64,     0,     0,     0,     0,     0,   217,   217,
       0,     0,    64,     0,     0,     0,     0,     0,     0,    65,
       0,     0,     0,     0,     0,    79,    80,    64,     0,     0,
     336,     0,     0,   188,    64,   235,    64,     0,     0,   336,
      96,     0,     0,     0,   546,    64,   235,    64,   546,   336,
     107,     0,   546,     0,     0,   217,     0,     0,     0,     0,
       0,     0,     0,     0,     0,   183,     0,     0,     0,   509,
     111,   190,   193,   258,   198,   113,     0,     0,   267,   509,
       0,     0,     0,   203,   119,   205,     0,     0,     0,   123,
     124,    69,     0,     0,     0,   217,   127,   113,     0,     0,
       0,     0,     0,   131,     0,     0,   119,     0,   235,   235,
       0,   123,   124,    69,     0,     0,   132,   133,   127,   134,
     135,     0,     0,   136,   217,   329,   137,     0,     0,     0,
       0,     0,     0,     0,     0,    64,   139,   140,   141,     0,
      70,    38,   214,    64,   235,     0,     0,     0,   330,     0,
       0,     0,     0,     0,   383,     0,     0,     0,     0,   235,
     141,     0,    70,    38,   384,     0,     0,   258,     0,     0,
       0,     0,   268,   321,     0,     0,     0,     0,     0,     0,
       0,   325,     0,     0,    54,     0,     0,     0,   604,   605,
       0,     0,     0,    55,     0,     0,     0,     0,    56,    57,
      58,     0,     0,     0,   217,    59,     0,     0,   258,     0,
       0,     0,    60,   217,     0,     0,     0,     0,   217,   111,
       0,     0,   217,   217,   113,     0,   217,     0,   509,    61,
       0,     0,     0,   119,     0,   509,   624,   509,   123,   124,
      69,    62,   340,     0,     0,   127,     0,    63,     0,     0,
      38,    64,   131,     0,     0,     0,     0,    64,     0,     0,
     188,    64,     0,     0,     0,   132,   133,     0,   134,   135,
       0,     0,   136,     0,     0,   137,     0,     0,     0,     0,
       0,   268,     0,     0,     0,   139,   140,   141,     0,    70,
      38,   660,   661,     0,   258,     0,     0,     0,   190,   444,
       0,    -2,   671,   509,   111,     0,     0,   112,     0,   113,
     114,   115,     0,     0,   116,     0,   117,   118,   672,   120,
     121,   122,     0,   123,   124,    69,   275,     0,   125,   126,
     127,   128,   129,   130,     0,     0,   283,   131,     0,     0,
     287,   288,   289,   290,   291,   292,   293,   294,   295,     0,
     132,   133,     0,   134,   135,     0,     0,   136,     0,     0,
     137,     0,     0,   -46,     0,     0,     0,     0,     0,     0,
     139,   140,   141,     0,    70,    38,   110,     0,   111,     0,
       0,   112,  -274,   113,   114,   115,     0,  -274,   116,     0,
     117,   118,   119,   120,   121,   122,     0,   123,   124,    69,
       0,     0,   125,   126,   127,   128,   129,   130,     0,     0,
       0,   131,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,   132,   133,     0,   134,   135,     0,
       0,   136,     0,     0,   137,     0,     0,  -274,     0,     0,
       0,     0,   138,  -274,   139,   140,   141,     0,    70,    38,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,   110,     0,   111,     0,     0,   112,     0,   113,   114,
     115,     0,    64,   116,    64,   117,   118,   119,   120,   121,
     122,     0,   123,   124,    69,     0,     0,   125,   126,   127,
     128,   129,   130,     0,     0,     0,   131,     0,     0,     0,
       0,     0,     0,     0,     0,     0,   273,  -299,     0,   132,
     133,     0,   134,   135,     0,     0,   136,     0,     0,   137,
       0,     0,  -274,     0,     0,     0,     0,   138,  -274,   139,
     140,   141,     0,    70,    38,   274,   275,     0,   276,   277,
     278,   279,   280,     0,   281,   282,   283,   284,   285,   286,
     287,   288,   289,   290,   291,   292,   293,   294,   295,   111,
       0,     0,     0,     0,   113,     0,     0,  -299,     0,     0,
       0,     0,     0,   119,     0,    54,     0,  -299,   123,   124,
      69,     0,     0,     0,    55,   127,     0,     0,     0,    56,
      57,    58,   256,     0,     0,     0,    59,     0,     0,     0,
       0,     0,     0,    60,     0,   132,   133,     0,   134,   257,
       0,     0,   136,     0,   111,   137,   418,     0,     0,   113,
      61,     0,     0,     0,     0,   139,   140,   141,   119,    70,
      38,     0,     0,   123,   124,    69,     0,     0,    63,     0,
     127,    38,     0,     0,     0,     0,     0,   131,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
     132,   133,     0,   134,   135,     0,     0,   136,     0,   111,
     137,     0,     0,     0,   113,     0,     0,     0,   508,     0,
     139,   140,   141,   119,    70,    38,     0,     0,   123,   124,
      69,     0,     0,     0,     0,   127,     0,   517,     0,     0,
       0,     0,   256,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,   132,   133,     0,   134,   257,
       0,     0,   136,     0,   111,   137,     0,     0,     0,   113,
       0,     0,     0,     0,     0,   139,   140,   141,   119,    70,
      38,     0,     0,   123,   124,    69,     0,   536,     0,     0,
     127,     0,     0,     0,     0,     0,     0,   131,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
     132,   133,     0,   134,   135,     0,     0,   136,     0,   111,
     137,     0,     0,     0,   113,     0,     0,     0,     0,     0,
     139,   140,   141,   119,    70,    38,     0,     0,   123,   124,
      69,     0,   537,     0,     0,   127,     0,     0,     0,     0,
       0,     0,   131,     0,     0,     0,     0,     0,     0,     0,
       0,     0,   111,     0,     0,   132,   133,   113,   134,   135,
       0,     0,   136,     0,     0,   137,   119,     0,     0,     0,
       0,   123,   124,    69,     0,   139,   140,   141,   127,    70,
      38,     0,     0,     0,     0,   131,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,   132,   133,
       0,   134,   135,     0,     0,   136,     0,   111,   137,     0,
       0,     0,   113,     0,     0,     0,   623,     0,   139,   140,
     141,   119,    70,    38,     0,     0,   123,   124,    69,     0,
       0,     0,     0,   127,     0,     0,     0,     0,     0,     0,
     131,     0,     0,     0,     0,     0,     0,     0,     0,     0,
     111,     0,     0,   132,   133,   113,   134,   135,     0,     0,
     136,     0,     0,   137,   119,     0,     0,     0,     0,   123,
     124,    69,     0,   139,   140,   141,   127,    70,    38,     0,
       0,     0,     0,   256,     0,     0,     0,     0,     0,     0,
       0,     0,     0,   111,     0,     0,   132,   133,   386,   134,
     257,     0,     0,   136,     0,     0,   137,   119,     0,     0,
       0,     0,   123,   124,    69,     0,   139,   140,   141,   127,
      70,    38,     0,     0,     0,     0,   131,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,   132,
     133,   113,   134,   135,     0,     0,   136,     0,     0,   137,
     119,     0,     0,     0,     0,   123,   124,    69,     0,   139,
     140,   141,   127,    70,    38,     0,   113,     0,     0,   329,
       0,     0,     0,     0,   113,   119,     0,   347,     0,     0,
     123,   124,    69,   119,     0,     0,   214,   127,   123,   124,
      69,     0,   330,     0,   329,   127,     0,     0,   394,     0,
     113,     0,   329,     0,   141,     0,    70,    38,   384,   119,
       0,   214,     0,     0,   123,   124,    69,   330,     0,   214,
       0,   127,     0,     0,     0,   330,     0,   113,   213,   141,
       0,    70,    38,   483,     0,     0,   119,   141,     0,    70,
      38,   123,   124,    69,     0,   214,     0,     0,   127,     0,
       0,   215,     0,   113,     0,   329,     0,     0,     0,     0,
       0,   113,   119,   141,     0,    70,    38,   123,   124,    69,
     119,     0,   214,     0,   127,   123,   124,    69,   330,     0,
       0,   493,   127,     0,     0,     0,   386,     0,     0,   329,
     141,     0,    70,    38,    54,   119,     0,     0,   214,     0,
     123,   124,    69,    55,   494,     0,   214,   127,    56,    57,
      58,     0,   540,     0,   329,    59,   141,     0,    70,    38,
      54,     0,    86,   106,   141,     0,    70,    38,     0,    55,
       0,   214,     0,     0,    56,    57,    58,   330,     0,    61,
       0,    59,     0,     0,     0,    87,     0,    54,    60,   141,
       0,    70,    38,     0,     0,     0,    55,    63,     0,     0,
      38,    56,    57,   186,     0,    61,     0,     0,    59,     0,
       0,     0,     0,    54,     0,    60,     0,     0,     0,     0,
       0,     0,    55,    63,     0,     0,    38,    56,    57,    58,
       0,     0,    61,     0,    59,     0,     0,     0,     0,     0,
       0,    60,     0,     0,     0,     0,     0,     0,     0,     0,
      63,   275,    70,    38,     0,   278,   279,   280,    61,     0,
     282,   283,   284,   285,   433,   287,   288,   289,   290,   291,
     292,   293,   294,   295,     0,     0,    63,   274,   275,    38,
     276,     0,   278,   279,   280,     0,     0,   282,   283,   284,
     285,   286,   287,   288,   289,   290,   291,   292,   293,   294,
     295,     0,     0,     0,     0,     0,     0,     0,     0,     0,
     274,   275,     0,   276,     0,   278,   279,   280,     0,   522,
     282,   283,   284,   285,   286,   287,   288,   289,   290,   291,
     292,   293,   294,   295,     0,     0,     0,     0,     0,     0,
     274,   275,     0,   276,     0,   278,   279,   280,     0,   519,
     282,   283,   284,   285,   286,   287,   288,   289,   290,   291,
     292,   293,   294,   295,   274,   275,     0,   276,     0,   278,
     279,   280,     0,   584,   282,   283,   284,   285,   286,   287,
     288,   289,   290,   291,   292,   293,   294,   295,   274,   275,
       0,   276,     0,   278,   279,   280,     0,   679,   282,   283,
     284,   285,   286,   287,   288,   289,   290,   291,   292,   293,
     294,   295,   274,   275,     0,   276,     0,   278,   279,   280,
       0,   680,   282,   283,   284,   285,   286,   287,   288,   289,
     290,   291,   292,   293,   294,   295,   274,   275,     0,     0,
       0,   278,   279,   280,     0,     0,   282,   283,   284,   285,
     286,   287,   288,   289,   290,   291,   292,   293,   294,   295,
     274,   275,     0,     0,     0,   278,   279,   280,     0,     0,
     282,   283,   284,   285,     0,   287,   288,   289,   290,   291,
     292,   293,   294,   295
};

static const yytype_int16 yycheck[] =
{
      11,   240,   264,    41,     6,   143,    17,   113,    10,   113,
     137,   226,   226,    24,    16,   299,   230,   205,   245,    91,
     137,    23,   316,   302,    62,    36,   240,   470,    99,   112,
     241,   143,   115,    35,   465,   233,   609,    75,   121,    77,
      51,   305,   304,   366,   367,   420,     3,    20,    50,    11,
      24,   426,     3,    91,    92,    35,    94,   127,   116,   427,
       3,    99,   120,    24,    59,   116,    35,   105,    63,   120,
     108,   109,    83,    24,   112,     5,    71,   115,     0,    59,
      82,    24,     5,   121,     5,    59,    25,    67,   370,    24,
     160,   129,   130,     1,    67,   165,   378,   540,    67,    50,
     382,    64,    62,    74,   113,   143,     3,    50,    59,     7,
      24,   393,    73,    75,    12,    72,   131,   132,   133,   134,
     135,   136,    73,    74,   139,   140,     3,    35,   137,    24,
      73,    74,   270,   321,   707,    65,    66,   512,   711,   514,
     257,     3,    65,     3,    65,    75,   125,    24,     8,    24,
     233,    59,    75,   476,    75,   478,   239,    17,   270,    67,
     358,   425,    22,    23,    24,    62,    63,    67,   179,    29,
      68,    68,    59,   665,     7,    60,   178,    66,    53,    12,
     297,    24,    59,    72,     3,   373,    63,    74,     1,   227,
      75,   213,   214,   215,    62,   233,    73,    74,    59,    59,
      68,   239,   694,   241,   213,   214,   215,   245,   576,    63,
      53,    71,   234,    73,    74,   583,    59,    59,   227,    67,
     242,   500,    35,     7,    71,   234,   264,   249,    12,   508,
      73,    74,   270,   242,   113,    68,   520,   531,    67,   533,
     249,   256,   257,   458,   458,   317,   689,   226,   257,     3,
     272,   466,   466,   324,    67,   698,   328,     7,   137,     8,
     298,   488,    12,   272,    24,   708,   304,   305,    65,   700,
      24,     3,    62,   535,   472,   486,    62,    24,    62,   317,
     386,   319,   386,   651,    68,   323,   324,    62,   297,   477,
     328,    35,    24,    68,   421,    62,   484,   579,    35,    59,
      62,   339,   313,    36,   421,    62,    60,    24,   330,    63,
     312,    68,    62,    73,    74,    59,     3,    62,    68,    73,
      74,   330,    59,    67,    24,   347,    73,    74,   307,   308,
      67,    63,    72,    62,   213,   214,   215,     3,   347,    68,
     349,    73,    74,    60,   623,   629,   384,   464,   227,    59,
     612,    35,   374,    71,    67,   234,    73,    74,   464,    59,
     464,   433,    35,   242,   386,   374,   504,    66,   390,   391,
     249,    66,    24,    73,    74,    59,   564,   386,   257,    35,
      59,   390,   391,    67,    59,    59,    59,   425,   515,   472,
      60,   606,   606,   272,    67,   433,    60,   480,   515,    75,
     483,    53,   684,    59,   383,    60,    59,    59,   113,    59,
      24,    67,   421,    59,    24,   394,    68,    28,   297,    30,
      31,    73,    74,    68,   462,   613,   614,   449,   439,   467,
      62,    72,   470,    68,   472,     9,   438,    62,    24,   556,
     449,    49,   480,    17,    24,   483,    60,    21,   486,    59,
     488,   330,   474,   462,   693,   464,    62,    31,    32,    73,
      74,   470,    59,    73,    74,   474,   504,    63,   347,    35,
     349,   493,   494,    59,    75,    60,     8,   456,   457,   693,
      60,   619,    68,    72,   493,   494,    24,    73,    74,    24,
      64,    72,   609,    73,    74,   374,    60,   535,    35,    24,
      62,    60,   540,   609,    75,   609,   515,   386,   213,   214,
     215,   390,   391,   492,    62,   642,    24,    59,   529,    62,
      72,    59,   227,    62,    59,   642,   528,    67,   507,   234,
      66,   540,    62,    62,    59,    73,    74,   242,    73,    74,
      60,    34,   421,    65,   249,    62,    65,   556,    73,    74,
      60,    44,    63,    60,    60,    48,    60,   125,   580,   597,
      53,    54,    55,    56,    60,    68,    68,   272,    35,   137,
     449,   580,    60,   141,   612,    60,    60,    75,    68,    60,
      75,   619,    60,   462,    60,   464,   703,    60,   599,    68,
     707,   470,    60,    25,   711,   474,    62,   703,    60,   703,
     609,   707,    66,   707,    60,   711,    60,   711,    60,    62,
      60,    60,    60,    68,   493,   494,   654,    68,   656,    62,
      21,    14,    62,    62,    68,   330,    72,   665,     3,     3,
      59,    62,    62,   642,   672,    60,   515,    60,    60,    59,
     656,    60,   347,   619,   349,   657,   349,   585,   585,   569,
     384,   689,   115,   349,   257,   691,   694,   668,   480,   462,
     698,   540,   472,   656,   556,   703,   668,    54,    54,   374,
     708,    87,   109,    41,   323,   686,   319,   556,   105,    -1,
     689,   386,    -1,    -1,    -1,   390,   391,    -1,    -1,   698,
      -1,    -1,    -1,    -1,   703,    -1,    -1,    -1,   707,   708,
      -1,   580,   711,    -1,    -1,   273,   274,   275,   276,    -1,
     278,   279,   280,    -1,   282,   283,   284,   285,   286,   287,
     288,   289,   290,   291,   292,   293,   294,   295,    -1,   297,
     609,   299,    -1,   301,    -1,    -1,    -1,    -1,    -1,   307,
     308,   309,    -1,    -1,   449,    -1,    -1,    -1,    27,    28,
      -1,    30,    31,    -1,    -1,    -1,    -1,   462,    -1,   464,
      -1,    40,    -1,   642,    -1,   470,    -1,    46,    47,   474,
      -1,    -1,   340,    -1,    -1,    54,    -1,    -1,    -1,    -1,
      -1,    -1,    61,    -1,    -1,    -1,    -1,    -1,   493,   494,
      -1,    -1,    71,    -1,    -1,    -1,    -1,    -1,    -1,    40,
      -1,    -1,    -1,    -1,    -1,    46,    47,    86,    -1,    -1,
     689,    -1,    -1,    92,    93,   383,    95,    -1,    -1,   698,
      61,    -1,    -1,    -1,   703,   104,   394,   106,   707,   708,
      71,    -1,   711,    -1,    -1,   540,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    86,    -1,    -1,    -1,   417,
       3,    92,    93,   421,    95,     8,    -1,    -1,    11,   427,
      -1,    -1,    -1,   104,    17,   106,    -1,    -1,    -1,    22,
      23,    24,    -1,    -1,    -1,   580,    29,     8,    -1,    -1,
      -1,    -1,    -1,    36,    -1,    -1,    17,    -1,   456,   457,
      -1,    22,    23,    24,    -1,    -1,    49,    50,    29,    52,
      53,    -1,    -1,    56,   609,    36,    59,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,   194,    69,    70,    71,    -1,
      73,    74,    53,   202,   492,    -1,    -1,    -1,    59,    -1,
      -1,    -1,    -1,    -1,    65,    -1,    -1,    -1,    -1,   507,
      71,    -1,    73,    74,    75,    -1,    -1,   515,    -1,    -1,
      -1,    -1,   520,   194,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,   202,    -1,    -1,     8,    -1,    -1,    -1,   536,   537,
      -1,    -1,    -1,    17,    -1,    -1,    -1,    -1,    22,    23,
      24,    -1,    -1,    -1,   689,    29,    -1,    -1,   556,    -1,
      -1,    -1,    36,   698,    -1,    -1,    -1,    -1,   703,     3,
      -1,    -1,   707,   708,     8,    -1,   711,    -1,   576,    53,
      -1,    -1,    -1,    17,    -1,   583,   584,   585,    22,    23,
      24,    65,    26,    -1,    -1,    29,    -1,    71,    -1,    -1,
      74,   310,    36,    -1,    -1,    -1,    -1,   316,    -1,    -1,
     319,   320,    -1,    -1,    -1,    49,    50,    -1,    52,    53,
      -1,    -1,    56,    -1,    -1,    59,    -1,    -1,    -1,    -1,
      -1,   629,    -1,    -1,    -1,    69,    70,    71,    -1,    73,
      74,   639,   640,    -1,   642,    -1,    -1,    -1,   319,   320,
      -1,     0,     1,   651,     3,    -1,    -1,     6,    -1,     8,
       9,    10,    -1,    -1,    13,    -1,    15,    16,    17,    18,
      19,    20,    -1,    22,    23,    24,    34,    -1,    27,    28,
      29,    30,    31,    32,    -1,    -1,    44,    36,    -1,    -1,
      48,    49,    50,    51,    52,    53,    54,    55,    56,    -1,
      49,    50,    -1,    52,    53,    -1,    -1,    56,    -1,    -1,
      59,    -1,    -1,    62,    -1,    -1,    -1,    -1,    -1,    -1,
      69,    70,    71,    -1,    73,    74,     1,    -1,     3,    -1,
      -1,     6,     7,     8,     9,    10,    -1,    12,    13,    -1,
      15,    16,    17,    18,    19,    20,    -1,    22,    23,    24,
      -1,    -1,    27,    28,    29,    30,    31,    32,    -1,    -1,
      -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    49,    50,    -1,    52,    53,    -1,
      -1,    56,    -1,    -1,    59,    -1,    -1,    62,    -1,    -1,
      -1,    -1,    67,    68,    69,    70,    71,    -1,    73,    74,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,     1,    -1,     3,    -1,    -1,     6,    -1,     8,     9,
      10,    -1,   531,    13,   533,    15,    16,    17,    18,    19,
      20,    -1,    22,    23,    24,    -1,    -1,    27,    28,    29,
      30,    31,    32,    -1,    -1,    -1,    36,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,     4,     5,    -1,    49,
      50,    -1,    52,    53,    -1,    -1,    56,    -1,    -1,    59,
      -1,    -1,    62,    -1,    -1,    -1,    -1,    67,    68,    69,
      70,    71,    -1,    73,    74,    33,    34,    -1,    36,    37,
      38,    39,    40,    -1,    42,    43,    44,    45,    46,    47,
      48,    49,    50,    51,    52,    53,    54,    55,    56,     3,
      -1,    -1,    -1,    -1,     8,    -1,    -1,    65,    -1,    -1,
      -1,    -1,    -1,    17,    -1,     8,    -1,    75,    22,    23,
      24,    -1,    -1,    -1,    17,    29,    -1,    -1,    -1,    22,
      23,    24,    36,    -1,    -1,    -1,    29,    -1,    -1,    -1,
      -1,    -1,    -1,    36,    -1,    49,    50,    -1,    52,    53,
      -1,    -1,    56,    -1,     3,    59,    60,    -1,    -1,     8,
      53,    -1,    -1,    -1,    -1,    69,    70,    71,    17,    73,
      74,    -1,    -1,    22,    23,    24,    -1,    -1,    71,    -1,
      29,    74,    -1,    -1,    -1,    -1,    -1,    36,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      49,    50,    -1,    52,    53,    -1,    -1,    56,    -1,     3,
      59,    -1,    -1,    -1,     8,    -1,    -1,    -1,    67,    -1,
      69,    70,    71,    17,    73,    74,    -1,    -1,    22,    23,
      24,    -1,    -1,    -1,    -1,    29,    -1,    31,    -1,    -1,
      -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    49,    50,    -1,    52,    53,
      -1,    -1,    56,    -1,     3,    59,    -1,    -1,    -1,     8,
      -1,    -1,    -1,    -1,    -1,    69,    70,    71,    17,    73,
      74,    -1,    -1,    22,    23,    24,    -1,    26,    -1,    -1,
      29,    -1,    -1,    -1,    -1,    -1,    -1,    36,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      49,    50,    -1,    52,    53,    -1,    -1,    56,    -1,     3,
      59,    -1,    -1,    -1,     8,    -1,    -1,    -1,    -1,    -1,
      69,    70,    71,    17,    73,    74,    -1,    -1,    22,    23,
      24,    -1,    26,    -1,    -1,    29,    -1,    -1,    -1,    -1,
      -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,     3,    -1,    -1,    49,    50,     8,    52,    53,
      -1,    -1,    56,    -1,    -1,    59,    17,    -1,    -1,    -1,
      -1,    22,    23,    24,    -1,    69,    70,    71,    29,    73,
      74,    -1,    -1,    -1,    -1,    36,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    49,    50,
      -1,    52,    53,    -1,    -1,    56,    -1,     3,    59,    -1,
      -1,    -1,     8,    -1,    -1,    -1,    67,    -1,    69,    70,
      71,    17,    73,    74,    -1,    -1,    22,    23,    24,    -1,
      -1,    -1,    -1,    29,    -1,    -1,    -1,    -1,    -1,    -1,
      36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
       3,    -1,    -1,    49,    50,     8,    52,    53,    -1,    -1,
      56,    -1,    -1,    59,    17,    -1,    -1,    -1,    -1,    22,
      23,    24,    -1,    69,    70,    71,    29,    73,    74,    -1,
      -1,    -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,     3,    -1,    -1,    49,    50,     8,    52,
      53,    -1,    -1,    56,    -1,    -1,    59,    17,    -1,    -1,
      -1,    -1,    22,    23,    24,    -1,    69,    70,    71,    29,
      73,    74,    -1,    -1,    -1,    -1,    36,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    49,
      50,     8,    52,    53,    -1,    -1,    56,    -1,    -1,    59,
      17,    -1,    -1,    -1,    -1,    22,    23,    24,    -1,    69,
      70,    71,    29,    73,    74,    -1,     8,    -1,    -1,    36,
      -1,    -1,    -1,    -1,     8,    17,    -1,    11,    -1,    -1,
      22,    23,    24,    17,    -1,    -1,    53,    29,    22,    23,
      24,    -1,    59,    -1,    36,    29,    -1,    -1,    65,    -1,
       8,    -1,    36,    -1,    71,    -1,    73,    74,    75,    17,
      -1,    53,    -1,    -1,    22,    23,    24,    59,    -1,    53,
      -1,    29,    -1,    -1,    -1,    59,    -1,     8,    36,    71,
      -1,    73,    74,    75,    -1,    -1,    17,    71,    -1,    73,
      74,    22,    23,    24,    -1,    53,    -1,    -1,    29,    -1,
      -1,    59,    -1,     8,    -1,    36,    -1,    -1,    -1,    -1,
      -1,     8,    17,    71,    -1,    73,    74,    22,    23,    24,
      17,    -1,    53,    -1,    29,    22,    23,    24,    59,    -1,
      -1,    36,    29,    -1,    -1,    -1,     8,    -1,    -1,    36,
      71,    -1,    73,    74,     8,    17,    -1,    -1,    53,    -1,
      22,    23,    24,    17,    59,    -1,    53,    29,    22,    23,
      24,    -1,    59,    -1,    36,    29,    71,    -1,    73,    74,
       8,    -1,    36,    11,    71,    -1,    73,    74,    -1,    17,
      -1,    53,    -1,    -1,    22,    23,    24,    59,    -1,    53,
      -1,    29,    -1,    -1,    -1,    59,    -1,     8,    36,    71,
      -1,    73,    74,    -1,    -1,    -1,    17,    71,    -1,    -1,
      74,    22,    23,    24,    -1,    53,    -1,    -1,    29,    -1,
      -1,    -1,    -1,     8,    -1,    36,    -1,    -1,    -1,    -1,
      -1,    -1,    17,    71,    -1,    -1,    74,    22,    23,    24,
      -1,    -1,    53,    -1,    29,    -1,    -1,    -1,    -1,    -1,
      -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      71,    34,    73,    74,    -1,    38,    39,    40,    53,    -1,
      43,    44,    45,    46,    59,    48,    49,    50,    51,    52,
      53,    54,    55,    56,    -1,    -1,    71,    33,    34,    74,
      36,    -1,    38,    39,    40,    -1,    -1,    43,    44,    45,
      46,    47,    48,    49,    50,    51,    52,    53,    54,    55,
      56,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      33,    34,    -1,    36,    -1,    38,    39,    40,    -1,    75,
      43,    44,    45,    46,    47,    48,    49,    50,    51,    52,
      53,    54,    55,    56,    -1,    -1,    -1,    -1,    -1,    -1,
      33,    34,    -1,    36,    -1,    38,    39,    40,    -1,    72,
      43,    44,    45,    46,    47,    48,    49,    50,    51,    52,
      53,    54,    55,    56,    33,    34,    -1,    36,    -1,    38,
      39,    40,    -1,    66,    43,    44,    45,    46,    47,    48,
      49,    50,    51,    52,    53,    54,    55,    56,    33,    34,
      -1,    36,    -1,    38,    39,    40,    -1,    66,    43,    44,
      45,    46,    47,    48,    49,    50,    51,    52,    53,    54,
      55,    56,    33,    34,    -1,    36,    -1,    38,    39,    40,
      -1,    66,    43,    44,    45,    46,    47,    48,    49,    50,
      51,    52,    53,    54,    55,    56,    33,    34,    -1,    -1,
      -1,    38,    39,    40,    -1,    -1,    43,    44,    45,    46,
      47,    48,    49,    50,    51,    52,    53,    54,    55,    56,
      33,    34,    -1,    -1,    -1,    38,    39,    40,    -1,    -1,
      43,    44,    45,    46,    -1,    48,    49,    50,    51,    52,
      53,    54,    55,    56
};

/* YYSTOS[STATE-NUM] -- The (internal number of the) accessing
   symbol of state STATE-NUM.  */
static const yytype_uint8 yystos[] =
{
       0,    77,    79,    80,     0,    99,   100,    25,   108,   101,
     102,   108,    24,   110,   111,    89,    90,   108,   110,    24,
     109,   236,    81,    82,   108,   110,    62,     9,    17,    21,
      31,    32,    64,   221,    91,    92,   108,   110,    74,   164,
     222,    59,   182,   222,    24,   222,   223,   222,    64,    93,
      94,   108,   110,     3,     8,    17,    22,    23,    24,    29,
      36,    53,    65,    71,   164,   224,   226,   227,   228,    24,
      73,   163,   164,   229,   237,    67,   184,    59,     3,   224,
     224,    97,    98,   108,   110,    63,    36,    59,   225,   226,
     228,    59,    67,    71,    67,     8,   224,     3,    50,    59,
     163,   234,   235,     3,    72,    65,    11,   224,    60,    75,
       1,     3,     6,     8,     9,    10,    13,    15,    16,    17,
      18,    19,    20,    22,    23,    27,    28,    29,    30,    31,
      32,    36,    49,    50,    52,    53,    56,    59,    67,    69,
      70,    71,   113,   114,   120,   122,   132,   135,   143,   146,
     148,   149,   150,   151,   156,   160,   163,   165,   166,   171,
     172,   175,   178,   179,   183,   186,   187,   202,   203,   205,
     208,    62,   217,   237,    62,    62,    62,    95,    96,   108,
     110,    24,    73,   224,   227,   217,    24,   163,   164,   219,
     224,   231,   239,   224,   163,   218,   230,   238,   224,     3,
     234,    62,    72,   224,   235,   224,     3,   220,   163,   229,
     160,   162,   163,    36,    53,    59,   163,   165,   170,   174,
     175,   176,   183,   162,   150,   156,   133,    59,   150,   160,
     136,    35,    67,   159,    71,   148,   208,   215,   147,   159,
     144,    59,   118,   119,   163,    59,   115,   161,   163,   207,
     149,   149,   149,   149,   149,   149,    36,    53,   148,   157,
     169,   175,   177,   183,   123,   149,   149,    11,   148,   214,
      59,   116,   207,     4,    33,    34,    36,    37,    38,    39,
      40,    42,    43,    44,    45,    46,    47,    48,    49,    50,
      51,    52,    53,    54,    55,    56,    67,    59,    63,    71,
      66,    59,   159,     1,   159,    62,    68,     5,    65,    75,
      60,    83,    84,   108,   110,    60,    60,    59,    68,    62,
      72,   224,    68,    62,    49,   224,    62,   220,    59,    36,
      59,   168,   174,   175,   176,   177,   183,   168,   168,    63,
      26,   120,   129,   130,   131,   208,   216,    11,   158,   163,
     167,   168,   199,   200,   201,   134,   216,    24,    59,    68,
     160,   193,   195,   197,   168,    35,    53,    59,    68,   160,
     192,   194,   195,   196,   206,   134,    60,   119,   191,   168,
      60,   115,   189,    65,    75,   168,     8,   169,    60,   205,
      72,    72,    60,   116,    65,   168,   148,   148,   148,   148,
     148,   148,   148,   148,   148,   148,   148,   148,   148,   148,
     148,   148,   148,   148,   148,   148,   148,   152,    60,   157,
     209,    59,   163,   148,   214,   204,   148,   152,   205,   202,
     208,   208,   148,    59,   224,   232,   233,    85,    86,   108,
     110,   232,   217,   231,   224,   220,   230,   234,   217,     8,
     168,    60,   163,   148,    35,   127,     5,    65,    62,   168,
     158,   167,    75,   213,    60,   137,    62,    63,    24,   195,
      59,   198,    62,   212,    72,   126,    59,   196,    53,   196,
      62,   212,   220,    75,   168,   145,    62,   212,    62,   212,
     208,   161,    65,    36,    59,   168,   174,   175,   176,   183,
      67,    68,   168,   168,    62,   212,   208,    65,    67,   148,
     153,   154,   210,   211,    11,    75,   213,    31,   157,    72,
      66,   202,    75,   213,   211,    68,   217,    87,    88,   108,
     110,    60,    60,    60,    60,   128,    26,    26,   216,   199,
      59,   173,   174,   175,   176,   177,   183,   185,   127,   216,
     163,    60,   201,   197,    68,   168,     7,    12,    68,   121,
     124,   196,   220,   196,    60,   194,    68,   160,   220,    35,
     119,    60,   115,    60,   208,   168,   152,   116,   117,   190,
     207,    60,   208,   152,    66,    75,   213,    68,   213,   157,
      60,    60,    60,   214,    60,    68,    60,    25,    78,   108,
     110,   232,   232,   205,   148,   148,    62,   201,   138,    60,
     209,    66,   125,    60,    60,   220,   126,    60,   211,    62,
     212,   168,   211,    67,   148,   155,   153,   154,    60,    66,
      72,   163,   103,   110,    68,   216,    60,   141,   185,     5,
      65,    66,    75,   205,   220,   220,    68,    68,   117,    60,
      68,   152,   214,    62,    21,   104,   188,    14,   139,   142,
     148,   148,   211,    72,     3,    59,    63,   105,   107,   163,
      62,     1,    17,   112,   113,   180,   203,    20,   122,    66,
      66,    68,    60,   105,   106,     3,   108,   110,     3,    59,
     163,   181,    62,   140,    62,   212,   110,   201,    59,   184,
     134,   105,    60,    60,   201,   127,   163,    60,    59,   185,
     201,    60,   185
};

#define yyerrok		(yyerrstatus = 0)
#define yyclearin	(yychar = YYEMPTY)
#define YYEMPTY		(-2)
#define YYEOF		0

#define YYACCEPT	goto yyacceptlab
#define YYABORT		goto yyabortlab
#define YYERROR		goto yyerrorlab


/* Like YYERROR except do call yyerror.  This remains here temporarily
   to ease the transition to the new meaning of YYERROR, for GCC.
   Once GCC version 2 has supplanted version 1, this can go.  */

#define YYFAIL		goto yyerrlab

#define YYRECOVERING()  (!!yyerrstatus)

#define YYBACKUP(Token, Value)					\
do								\
  if (yychar == YYEMPTY && yylen == 1)				\
    {								\
      yychar = (Token);						\
      yylval = (Value);						\
      yytoken = YYTRANSLATE (yychar);				\
      YYPOPSTACK (1);						\
      goto yybackup;						\
    }								\
  else								\
    {								\
      yyerror (YY_("syntax error: cannot back up")); \
      YYERROR;							\
    }								\
while (YYID (0))


#define YYTERROR	1
#define YYERRCODE	256


/* YYLLOC_DEFAULT -- Set CURRENT to span from RHS[1] to RHS[N].
   If N is 0, then set CURRENT to the empty location which ends
   the previous symbol: RHS[0] (always defined).  */

#define YYRHSLOC(Rhs, K) ((Rhs)[K])
#ifndef YYLLOC_DEFAULT
# define YYLLOC_DEFAULT(Current, Rhs, N)				\
    do									\
      if (YYID (N))                                                    \
	{								\
	  (Current).first_line   = YYRHSLOC (Rhs, 1).first_line;	\
	  (Current).first_column = YYRHSLOC (Rhs, 1).first_column;	\
	  (Current).last_line    = YYRHSLOC (Rhs, N).last_line;		\
	  (Current).last_column  = YYRHSLOC (Rhs, N).last_column;	\
	}								\
      else								\
	{								\
	  (Current).first_line   = (Current).last_line   =		\
	    YYRHSLOC (Rhs, 0).last_line;				\
	  (Current).first_column = (Current).last_column =		\
	    YYRHSLOC (Rhs, 0).last_column;				\
	}								\
    while (YYID (0))
#endif


/* YY_LOCATION_PRINT -- Print the location on the stream.
   This macro was not mandated originally: define only if we know
   we won't break user code: when these are the locations we know.  */

#ifndef YY_LOCATION_PRINT
# if defined YYLTYPE_IS_TRIVIAL && YYLTYPE_IS_TRIVIAL
#  define YY_LOCATION_PRINT(File, Loc)			\
     fprintf (File, "%d.%d-%d.%d",			\
	      (Loc).first_line, (Loc).first_column,	\
	      (Loc).last_line,  (Loc).last_column)
# else
#  define YY_LOCATION_PRINT(File, Loc) ((void) 0)
# endif
#endif


/* YYLEX -- calling `yylex' with the right arguments.  */

#ifdef YYLEX_PARAM
# define YYLEX yylex (YYLEX_PARAM)
#else
# define YYLEX yylex ()
#endif

/* Enable debugging if requested.  */
#if YYDEBUG

# ifndef YYFPRINTF
#  include <stdio.h> /* INFRINGES ON USER NAME SPACE */
#  define YYFPRINTF fprintf
# endif

# define YYDPRINTF(Args)			\
do {						\
  if (yydebug)					\
    YYFPRINTF Args;				\
} while (YYID (0))

# define YY_SYMBOL_PRINT(Title, Type, Value, Location)			  \
do {									  \
  if (yydebug)								  \
    {									  \
      YYFPRINTF (stderr, "%s ", Title);					  \
      yy_symbol_print (stderr,						  \
		  Type, Value); \
      YYFPRINTF (stderr, "\n");						  \
    }									  \
} while (YYID (0))


/*--------------------------------.
| Print this symbol on YYOUTPUT.  |
`--------------------------------*/

/*ARGSUSED*/
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yy_symbol_value_print (FILE *yyoutput, int yytype, YYSTYPE const * const yyvaluep)
#else
static void
yy_symbol_value_print (yyoutput, yytype, yyvaluep)
    FILE *yyoutput;
    int yytype;
    YYSTYPE const * const yyvaluep;
#endif
{
  if (!yyvaluep)
    return;
# ifdef YYPRINT
  if (yytype < YYNTOKENS)
    YYPRINT (yyoutput, yytoknum[yytype], *yyvaluep);
# else
  YYUSE (yyoutput);
# endif
  switch (yytype)
    {
      default:
	break;
    }
}


/*--------------------------------.
| Print this symbol on YYOUTPUT.  |
`--------------------------------*/

#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yy_symbol_print (FILE *yyoutput, int yytype, YYSTYPE const * const yyvaluep)
#else
static void
yy_symbol_print (yyoutput, yytype, yyvaluep)
    FILE *yyoutput;
    int yytype;
    YYSTYPE const * const yyvaluep;
#endif
{
  if (yytype < YYNTOKENS)
    YYFPRINTF (yyoutput, "token %s (", yytname[yytype]);
  else
    YYFPRINTF (yyoutput, "nterm %s (", yytname[yytype]);

  yy_symbol_value_print (yyoutput, yytype, yyvaluep);
  YYFPRINTF (yyoutput, ")");
}

/*------------------------------------------------------------------.
| yy_stack_print -- Print the state stack from its BOTTOM up to its |
| TOP (included).                                                   |
`------------------------------------------------------------------*/

#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yy_stack_print (yytype_int16 *bottom, yytype_int16 *top)
#else
static void
yy_stack_print (bottom, top)
    yytype_int16 *bottom;
    yytype_int16 *top;
#endif
{
  YYFPRINTF (stderr, "Stack now");
  for (; bottom <= top; ++bottom)
    YYFPRINTF (stderr, " %d", *bottom);
  YYFPRINTF (stderr, "\n");
}

# define YY_STACK_PRINT(Bottom, Top)				\
do {								\
  if (yydebug)							\
    yy_stack_print ((Bottom), (Top));				\
} while (YYID (0))


/*------------------------------------------------.
| Report that the YYRULE is going to be reduced.  |
`------------------------------------------------*/

#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yy_reduce_print (YYSTYPE *yyvsp, int yyrule)
#else
static void
yy_reduce_print (yyvsp, yyrule)
    YYSTYPE *yyvsp;
    int yyrule;
#endif
{
  int yynrhs = yyr2[yyrule];
  int yyi;
  unsigned long int yylno = yyrline[yyrule];
  YYFPRINTF (stderr, "Reducing stack by rule %d (line %lu):\n",
	     yyrule - 1, yylno);
  /* The symbols being reduced.  */
  for (yyi = 0; yyi < yynrhs; yyi++)
    {
      fprintf (stderr, "   $%d = ", yyi + 1);
      yy_symbol_print (stderr, yyrhs[yyprhs[yyrule] + yyi],
		       &(yyvsp[(yyi + 1) - (yynrhs)])
		       		       );
      fprintf (stderr, "\n");
    }
}

# define YY_REDUCE_PRINT(Rule)		\
do {					\
  if (yydebug)				\
    yy_reduce_print (yyvsp, Rule); \
} while (YYID (0))

/* Nonzero means print parse trace.  It is left uninitialized so that
   multiple parsers can coexist.  */
int yydebug;
#else /* !YYDEBUG */
# define YYDPRINTF(Args)
# define YY_SYMBOL_PRINT(Title, Type, Value, Location)
# define YY_STACK_PRINT(Bottom, Top)
# define YY_REDUCE_PRINT(Rule)
#endif /* !YYDEBUG */


/* YYINITDEPTH -- initial size of the parser's stacks.  */
#ifndef	YYINITDEPTH
# define YYINITDEPTH 200
#endif

/* YYMAXDEPTH -- maximum size the stacks can grow to (effective only
   if the built-in stack extension method is used).

   Do not make this value too large; the results are undefined if
   YYSTACK_ALLOC_MAXIMUM < YYSTACK_BYTES (YYMAXDEPTH)
   evaluated with infinite-precision integer arithmetic.  */

#ifndef YYMAXDEPTH
# define YYMAXDEPTH 10000
#endif



#if YYERROR_VERBOSE

# ifndef yystrlen
#  if defined __GLIBC__ && defined _STRING_H
#   define yystrlen strlen
#  else
/* Return the length of YYSTR.  */
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static YYSIZE_T
yystrlen (const char *yystr)
#else
static YYSIZE_T
yystrlen (yystr)
    const char *yystr;
#endif
{
  YYSIZE_T yylen;
  for (yylen = 0; yystr[yylen]; yylen++)
    continue;
  return yylen;
}
#  endif
# endif

# ifndef yystpcpy
#  if defined __GLIBC__ && defined _STRING_H && defined _GNU_SOURCE
#   define yystpcpy stpcpy
#  else
/* Copy YYSRC to YYDEST, returning the address of the terminating '\0' in
   YYDEST.  */
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static char *
yystpcpy (char *yydest, const char *yysrc)
#else
static char *
yystpcpy (yydest, yysrc)
    char *yydest;
    const char *yysrc;
#endif
{
  char *yyd = yydest;
  const char *yys = yysrc;

  while ((*yyd++ = *yys++) != '\0')
    continue;

  return yyd - 1;
}
#  endif
# endif

# ifndef yytnamerr
/* Copy to YYRES the contents of YYSTR after stripping away unnecessary
   quotes and backslashes, so that it's suitable for yyerror.  The
   heuristic is that double-quoting is unnecessary unless the string
   contains an apostrophe, a comma, or backslash (other than
   backslash-backslash).  YYSTR is taken from yytname.  If YYRES is
   null, do not copy; instead, return the length of what the result
   would have been.  */
static YYSIZE_T
yytnamerr (char *yyres, const char *yystr)
{
  if (*yystr == '"')
    {
      YYSIZE_T yyn = 0;
      char const *yyp = yystr;

      for (;;)
	switch (*++yyp)
	  {
	  case '\'':
	  case ',':
	    goto do_not_strip_quotes;

	  case '\\':
	    if (*++yyp != '\\')
	      goto do_not_strip_quotes;
	    /* Fall through.  */
	  default:
	    if (yyres)
	      yyres[yyn] = *yyp;
	    yyn++;
	    break;

	  case '"':
	    if (yyres)
	      yyres[yyn] = '\0';
	    return yyn;
	  }
    do_not_strip_quotes: ;
    }

  if (! yyres)
    return yystrlen (yystr);

  return yystpcpy (yyres, yystr) - yyres;
}
# endif

/* Copy into YYRESULT an error message about the unexpected token
   YYCHAR while in state YYSTATE.  Return the number of bytes copied,
   including the terminating null byte.  If YYRESULT is null, do not
   copy anything; just return the number of bytes that would be
   copied.  As a special case, return 0 if an ordinary "syntax error"
   message will do.  Return YYSIZE_MAXIMUM if overflow occurs during
   size calculation.  */
static YYSIZE_T
yysyntax_error (char *yyresult, int yystate, int yychar)
{
  int yyn = yypact[yystate];

  if (! (YYPACT_NINF < yyn && yyn <= YYLAST))
    return 0;
  else
    {
      int yytype = YYTRANSLATE (yychar);
      YYSIZE_T yysize0 = yytnamerr (0, yytname[yytype]);
      YYSIZE_T yysize = yysize0;
      YYSIZE_T yysize1;
      int yysize_overflow = 0;
      enum { YYERROR_VERBOSE_ARGS_MAXIMUM = 5 };
      char const *yyarg[YYERROR_VERBOSE_ARGS_MAXIMUM];
      int yyx;

# if 0
      /* This is so xgettext sees the translatable formats that are
	 constructed on the fly.  */
      YY_("syntax error, unexpected %s");
      YY_("syntax error, unexpected %s, expecting %s");
      YY_("syntax error, unexpected %s, expecting %s or %s");
      YY_("syntax error, unexpected %s, expecting %s or %s or %s");
      YY_("syntax error, unexpected %s, expecting %s or %s or %s or %s");
# endif
      char *yyfmt;
      char const *yyf;
      static char const yyunexpected[] = "syntax error, unexpected %s";
      static char const yyexpecting[] = ", expecting %s";
      static char const yyor[] = " or %s";
      char yyformat[sizeof yyunexpected
		    + sizeof yyexpecting - 1
		    + ((YYERROR_VERBOSE_ARGS_MAXIMUM - 2)
		       * (sizeof yyor - 1))];
      char const *yyprefix = yyexpecting;

      /* Start YYX at -YYN if negative to avoid negative indexes in
	 YYCHECK.  */
      int yyxbegin = yyn < 0 ? -yyn : 0;

      /* Stay within bounds of both yycheck and yytname.  */
      int yychecklim = YYLAST - yyn + 1;
      int yyxend = yychecklim < YYNTOKENS ? yychecklim : YYNTOKENS;
      int yycount = 1;

      yyarg[0] = yytname[yytype];
      yyfmt = yystpcpy (yyformat, yyunexpected);

      for (yyx = yyxbegin; yyx < yyxend; ++yyx)
	if (yycheck[yyx + yyn] == yyx && yyx != YYTERROR)
	  {
	    if (yycount == YYERROR_VERBOSE_ARGS_MAXIMUM)
	      {
		yycount = 1;
		yysize = yysize0;
		yyformat[sizeof yyunexpected - 1] = '\0';
		break;
	      }
	    yyarg[yycount++] = yytname[yyx];
	    yysize1 = yysize + yytnamerr (0, yytname[yyx]);
	    yysize_overflow |= (yysize1 < yysize);
	    yysize = yysize1;
	    yyfmt = yystpcpy (yyfmt, yyprefix);
	    yyprefix = yyor;
	  }

      yyf = YY_(yyformat);
      yysize1 = yysize + yystrlen (yyf);
      yysize_overflow |= (yysize1 < yysize);
      yysize = yysize1;

      if (yysize_overflow)
	return YYSIZE_MAXIMUM;

      if (yyresult)
	{
	  /* Avoid sprintf, as that infringes on the user's name space.
	     Don't have undefined behavior even if the translation
	     produced a string with the wrong number of "%s"s.  */
	  char *yyp = yyresult;
	  int yyi = 0;
	  while ((*yyp = *yyf) != '\0')
	    {
	      if (*yyp == '%' && yyf[1] == 's' && yyi < yycount)
		{
		  yyp += yytnamerr (yyp, yyarg[yyi++]);
		  yyf += 2;
		}
	      else
		{
		  yyp++;
		  yyf++;
		}
	    }
	}
      return yysize;
    }
}
#endif /* YYERROR_VERBOSE */


/*-----------------------------------------------.
| Release the memory associated to this symbol.  |
`-----------------------------------------------*/

/*ARGSUSED*/
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yydestruct (const char *yymsg, int yytype, YYSTYPE *yyvaluep)
#else
static void
yydestruct (yymsg, yytype, yyvaluep)
    const char *yymsg;
    int yytype;
    YYSTYPE *yyvaluep;
#endif
{
  YYUSE (yyvaluep);

  if (!yymsg)
    yymsg = "Deleting";
  YY_SYMBOL_PRINT (yymsg, yytype, yyvaluep, yylocationp);

  switch (yytype)
    {

      default:
	break;
    }
}


/* Prevent warnings from -Wmissing-prototypes.  */

#ifdef YYPARSE_PARAM
#if defined __STDC__ || defined __cplusplus
int yyparse (void *YYPARSE_PARAM);
#else
int yyparse ();
#endif
#else /* ! YYPARSE_PARAM */
#if defined __STDC__ || defined __cplusplus
int yyparse (void);
#else
int yyparse ();
#endif
#endif /* ! YYPARSE_PARAM */



/* The look-ahead symbol.  */
int yychar, yystate;

/* The semantic value of the look-ahead symbol.  */
YYSTYPE yylval;

/* Number of syntax errors so far.  */
int yynerrs;



/*----------.
| yyparse.  |
`----------*/

#ifdef YYPARSE_PARAM
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
int
yyparse (void *YYPARSE_PARAM)
#else
int
yyparse (YYPARSE_PARAM)
    void *YYPARSE_PARAM;
#endif
#else /* ! YYPARSE_PARAM */
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
int
yyparse (void)
#else
int
yyparse ()

#endif
#endif
{
  
  int yyn;
  int yyresult;
  /* Number of tokens to shift before error messages enabled.  */
  int yyerrstatus;
  /* Look-ahead token as an internal (translated) token number.  */
  int yytoken = 0;
#if YYERROR_VERBOSE
  /* Buffer for error messages, and its allocated size.  */
  char yymsgbuf[128];
  char *yymsg = yymsgbuf;
  YYSIZE_T yymsg_alloc = sizeof yymsgbuf;
#endif

  /* Three stacks and their tools:
     `yyss': related to states,
     `yyvs': related to semantic values,
     `yyls': related to locations.

     Refer to the stacks thru separate pointers, to allow yyoverflow
     to reallocate them elsewhere.  */

  /* The state stack.  */
  yytype_int16 yyssa[YYINITDEPTH];
  yytype_int16 *yyss = yyssa;
  yytype_int16 *yyssp;

  /* The semantic value stack.  */
  YYSTYPE yyvsa[YYINITDEPTH];
  YYSTYPE *yyvs = yyvsa;
  YYSTYPE *yyvsp;



#define YYPOPSTACK(N)   (yyvsp -= (N), yyssp -= (N))

  YYSIZE_T yystacksize = YYINITDEPTH;

  /* The variables used to return semantic value and location from the
     action routines.  */
  YYSTYPE yyval;


  /* The number of symbols on the RHS of the reduced rule.
     Keep to zero when no symbol should be popped.  */
  int yylen = 0;

  YYDPRINTF ((stderr, "Starting parse\n"));

  yystate = 0;
  yyerrstatus = 0;
  yynerrs = 0;
  yychar = YYEMPTY;		/* Cause a token to be read.  */

  /* Initialize stack pointers.
     Waste one element of value and location stack
     so that they stay on the same level as the state stack.
     The wasted elements are never initialized.  */

  yyssp = yyss;
  yyvsp = yyvs;

  goto yysetstate;

/*------------------------------------------------------------.
| yynewstate -- Push a new state, which is found in yystate.  |
`------------------------------------------------------------*/
 yynewstate:
  /* In all cases, when you get here, the value and location stacks
     have just been pushed.  So pushing a state here evens the stacks.  */
  yyssp++;

 yysetstate:
  *yyssp = yystate;

  if (yyss + yystacksize - 1 <= yyssp)
    {
      /* Get the current used size of the three stacks, in elements.  */
      YYSIZE_T yysize = yyssp - yyss + 1;

#ifdef yyoverflow
      {
	/* Give user a chance to reallocate the stack.  Use copies of
	   these so that the &'s don't force the real ones into
	   memory.  */
	YYSTYPE *yyvs1 = yyvs;
	yytype_int16 *yyss1 = yyss;


	/* Each stack pointer address is followed by the size of the
	   data in use in that stack, in bytes.  This used to be a
	   conditional around just the two extra args, but that might
	   be undefined if yyoverflow is a macro.  */
	yyoverflow (YY_("memory exhausted"),
		    &yyss1, yysize * sizeof (*yyssp),
		    &yyvs1, yysize * sizeof (*yyvsp),

		    &yystacksize);

	yyss = yyss1;
	yyvs = yyvs1;
      }
#else /* no yyoverflow */
# ifndef YYSTACK_RELOCATE
      goto yyexhaustedlab;
# else
      /* Extend the stack our own way.  */
      if (YYMAXDEPTH <= yystacksize)
	goto yyexhaustedlab;
      yystacksize *= 2;
      if (YYMAXDEPTH < yystacksize)
	yystacksize = YYMAXDEPTH;

      {
	yytype_int16 *yyss1 = yyss;
	union yyalloc *yyptr =
	  (union yyalloc *) YYSTACK_ALLOC (YYSTACK_BYTES (yystacksize));
	if (! yyptr)
	  goto yyexhaustedlab;
	YYSTACK_RELOCATE (yyss);
	YYSTACK_RELOCATE (yyvs);

#  undef YYSTACK_RELOCATE
	if (yyss1 != yyssa)
	  YYSTACK_FREE (yyss1);
      }
# endif
#endif /* no yyoverflow */

      yyssp = yyss + yysize - 1;
      yyvsp = yyvs + yysize - 1;


      YYDPRINTF ((stderr, "Stack size increased to %lu\n",
		  (unsigned long int) yystacksize));

      if (yyss + yystacksize - 1 <= yyssp)
	YYABORT;
    }

  YYDPRINTF ((stderr, "Entering state %d\n", yystate));

  goto yybackup;

/*-----------.
| yybackup.  |
`-----------*/
yybackup:

  /* Do appropriate processing given the current state.  Read a
     look-ahead token if we need one and don't already have one.  */

  /* First try to decide what to do without reference to look-ahead token.  */
  yyn = yypact[yystate];
  if (yyn == YYPACT_NINF)
    goto yydefault;

  /* Not known => get a look-ahead token if don't already have one.  */

  /* YYCHAR is either YYEMPTY or YYEOF or a valid look-ahead symbol.  */
  if (yychar == YYEMPTY)
    {
      YYDPRINTF ((stderr, "Reading a token: "));
      yychar = YYLEX;
    }

  if (yychar <= YYEOF)
    {
      yychar = yytoken = YYEOF;
      YYDPRINTF ((stderr, "Now at end of input.\n"));
    }
  else
    {
      yytoken = YYTRANSLATE (yychar);
      YY_SYMBOL_PRINT ("Next token is", yytoken, &yylval, &yylloc);
    }

  /* If the proper action on seeing token YYTOKEN is to reduce or to
     detect an error, take that action.  */
  yyn += yytoken;
  if (yyn < 0 || YYLAST < yyn || yycheck[yyn] != yytoken)
    goto yydefault;
  yyn = yytable[yyn];
  if (yyn <= 0)
    {
      if (yyn == 0 || yyn == YYTABLE_NINF)
	goto yyerrlab;
      yyn = -yyn;
      goto yyreduce;
    }

  if (yyn == YYFINAL)
    YYACCEPT;

  /* Count tokens shifted since error; after three, turn off error
     status.  */
  if (yyerrstatus)
    yyerrstatus--;

  /* Shift the look-ahead token.  */
  YY_SYMBOL_PRINT ("Shifting", yytoken, &yylval, &yylloc);

  /* Discard the shifted token unless it is eof.  */
  if (yychar != YYEOF)
    yychar = YYEMPTY;

  yystate = yyn;
  *++yyvsp = yylval;

  goto yynewstate;


/*-----------------------------------------------------------.
| yydefault -- do the default action for the current state.  |
`-----------------------------------------------------------*/
yydefault:
  yyn = yydefact[yystate];
  if (yyn == 0)
    goto yyerrlab;
  goto yyreduce;


/*-----------------------------.
| yyreduce -- Do a reduction.  |
`-----------------------------*/
yyreduce:
  /* yyn is the number of a rule to reduce with.  */
  yylen = yyr2[yyn];

  /* If YYLEN is nonzero, implement the default value of the action:
     `$$ = $1'.

     Otherwise, the following line sets YYVAL to garbage.
     This behavior is undocumented and Bison
     users should not rely upon it.  Assigning to YYVAL
     unconditionally makes the parser a bit smaller, and it avoids a
     GCC warning that YYVAL may be used uninitialized.  */
  yyval = yyvsp[1-yylen];


  YY_REDUCE_PRINT (yyn);
  switch (yyn)
    {
        case 2:
#line 139 "go.y"
    {
		xtop = concat(xtop, (yyvsp[(15) - (15)].list));
	}
    break;

  case 3:
#line 145 "go.y"
    {
		prevlineno = lineno;
		yyerror("package statement must be first");
		errorexit();
	}
    break;

  case 4:
#line 151 "go.y"
    {
		mkpackage((yyvsp[(2) - (3)].sym)->name);
	}
    break;

  case 5:
#line 161 "go.y"
    {
		importpkg = corepkg;
		if(debug['A']) {
			cannedimports("core.builtin", "package core\n\n$$\n\n");
		} else {
			cannedimports("core.builtin", coreimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 6:
#line 173 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 7:
#line 178 "go.y"
    {
		importpkg = channelspkg;
		if(debug['A']) {
			cannedimports("channels.builtin", "package channels\n\n$$\n\n");
		} else {
			cannedimports("channels.builtin", channelsimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 8:
#line 190 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 9:
#line 196 "go.y"
    {
		importpkg = seqpkg;
		if(debug['A']) {
			cannedimports("seq.builtin", "package seq\n\n$$\n\n");
		} else {
			cannedimports("seq.builtin", seqimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 10:
#line 208 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 11:
#line 214 "go.y"
    {
		importpkg = deferspkg;
		if(debug['A']) {
			cannedimports("defers.builtin", "package defers\n\n$$\n\n");
		} else {
			cannedimports("defers.builtin", defersimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 12:
#line 226 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 13:
#line 232 "go.y"
    {
		importpkg = runtimepkg;
		if(debug['A']) {
			cannedimports("runtime.builtin", "package runtime\n\n$$\n\n");
		} else {
			cannedimports("runtime.builtin", runtimeimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 14:
#line 244 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 15:
#line 250 "go.y"
    {
		importpkg = schedpkg;
		if(debug['A']) {
			cannedimports("sched.builtin", "package sched\n\n$$\n\n");
		} else {
			cannedimports("sched.builtin", schedimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 16:
#line 262 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 17:
#line 268 "go.y"
    {
		importpkg = hashpkg;
		if(debug['A']) {
			cannedimports("hash.builtin", "package hash\n\n$$\n\n");
		} else {
			cannedimports("hash.builtin", hashimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 18:
#line 280 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 19:
#line 285 "go.y"
    {
		importpkg = printfpkg;
		if(debug['A']) {
			cannedimports("printf.builtin", "package printf\n\n$$\n\n");
		} else {
			cannedimports("printf.builtin", printfimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 20:
#line 297 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 21:
#line 303 "go.y"
    {
		importpkg = ifacestuffpkg;
		if(debug['A']) {
			cannedimports("ifacestuff.builtin", "package ifacestuff\n\n$$\n\n");
		} else {
			cannedimports("ifacestuff.builtin", ifacestuffimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 22:
#line 315 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 23:
#line 321 "go.y"
    {
		importpkg = stringspkg;
		if(debug['A']) {
			cannedimports("strings.builtin", "package strings\n\n$$\n\n");
		} else {
			cannedimports("strings.builtin", stringsimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 24:
#line 333 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 25:
#line 339 "go.y"
    {
		importpkg = mapspkg;
		if(debug['A']) {
			cannedimports("maps.builtin", "package maps\n\n$$\n\n");
		} else {
			cannedimports("maps.builtin", mapsimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 26:
#line 351 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 27:
#line 362 "go.y"
    {
		importpkg = wbpkg;
			if(debug['A']) {
				cannedimports("stackwb.builtin", "package stackwb\n\n$$\n\n");
			} else {
				cannedimports("stackwb.builtin", wbimport);
			}
		curio.importsafe = 1;
	}
    break;

  case 28:
#line 373 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 34:
#line 388 "go.y"
    {
		Pkg *ipkg;
		Sym *my;
		Node *pack;
		
		ipkg = importpkg;
		my = importmyname;
		importpkg = nil;
		importmyname = S;

		if(my == nil)
			my = lookup(ipkg->name);

		pack = nod(OPACK, N, N);
		pack->sym = my;
		pack->pkg = ipkg;
		pack->lineno = (yyvsp[(1) - (3)].i);

		if(my->name[0] == '.') {
			importdot(ipkg, pack);
			break;
		}
		if(strcmp(my->name, "init") == 0) {
			yyerror("cannot import package as init - init must be a func");
			break;
		}
		if(my->name[0] == '_' && my->name[1] == '\0')
			break;
		if(my->def) {
			lineno = (yyvsp[(1) - (3)].i);
			redeclare(my, "as imported package name");
		}
		my->def = pack;
		my->lastlineno = (yyvsp[(1) - (3)].i);
		my->block = 1;	// at top level
	}
    break;

  case 35:
#line 425 "go.y"
    {
		// When an invalid import path is passed to importfile,
		// it calls yyerror and then sets up a fake import with
		// no package statement. This allows us to test more
		// than one invalid import statement in a single file.
		if(nerrors == 0)
			fatal("phase error in import");
	}
    break;

  case 38:
#line 440 "go.y"
    {
		// import with original name
		(yyval.i) = parserline();
		importmyname = S;
		importfile(&(yyvsp[(1) - (1)].val), (yyval.i));
	}
    break;

  case 39:
#line 447 "go.y"
    {
		// import with given name
		(yyval.i) = parserline();
		importmyname = (yyvsp[(1) - (2)].sym);
		importfile(&(yyvsp[(2) - (2)].val), (yyval.i));
	}
    break;

  case 40:
#line 454 "go.y"
    {
		// import into my name space
		(yyval.i) = parserline();
		importmyname = lookup(".");
		importfile(&(yyvsp[(2) - (2)].val), (yyval.i));
	}
    break;

  case 41:
#line 463 "go.y"
    {
		if(importpkg->name == nil) {
			importpkg->name = (yyvsp[(2) - (4)].sym)->name;
			pkglookup((yyvsp[(2) - (4)].sym)->name, nil)->npkg++;
		} else if(strcmp(importpkg->name, (yyvsp[(2) - (4)].sym)->name) != 0)
			yyerror("conflicting names %s and %s for package \"%Z\"", importpkg->name, (yyvsp[(2) - (4)].sym)->name, importpkg->path);
		importpkg->direct = 1;
		importpkg->safe = curio.importsafe;

		if(safemode && !curio.importsafe)
			yyerror("cannot import unsafe package \"%Z\"", importpkg->path);
	}
    break;

  case 43:
#line 478 "go.y"
    {
		if(strcmp((yyvsp[(1) - (1)].sym)->name, "safe") == 0)
			curio.importsafe = 1;
	}
    break;

  case 44:
#line 484 "go.y"
    {
		defercheckwidth();
	}
    break;

  case 45:
#line 488 "go.y"
    {
		resumecheckwidth();
		unimportfile();
	}
    break;

  case 46:
#line 497 "go.y"
    {
		yyerror("empty top-level declaration");
		(yyval.list) = nil;
	}
    break;

  case 48:
#line 503 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 49:
#line 507 "go.y"
    {
		yyerror("non-declaration statement outside function body");
		(yyval.list) = nil;
	}
    break;

  case 50:
#line 512 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 51:
#line 518 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (2)].list);
	}
    break;

  case 52:
#line 522 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (5)].list);
	}
    break;

  case 53:
#line 526 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 54:
#line 530 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (2)].list);
		iota = -100000;
		lastconst = nil;
	}
    break;

  case 55:
#line 536 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (5)].list);
		iota = -100000;
		lastconst = nil;
	}
    break;

  case 56:
#line 542 "go.y"
    {
		(yyval.list) = concat((yyvsp[(3) - (7)].list), (yyvsp[(5) - (7)].list));
		iota = -100000;
		lastconst = nil;
	}
    break;

  case 57:
#line 548 "go.y"
    {
		(yyval.list) = nil;
		iota = -100000;
	}
    break;

  case 58:
#line 553 "go.y"
    {
		(yyval.list) = list1((yyvsp[(2) - (2)].node));
	}
    break;

  case 59:
#line 557 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (5)].list);
	}
    break;

  case 60:
#line 561 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 61:
#line 567 "go.y"
    {
		iota = 0;
	}
    break;

  case 62:
#line 573 "go.y"
    {
		(yyval.list) = variter((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].node), nil);
	}
    break;

  case 63:
#line 577 "go.y"
    {
		(yyval.list) = variter((yyvsp[(1) - (4)].list), (yyvsp[(2) - (4)].node), (yyvsp[(4) - (4)].list));
	}
    break;

  case 64:
#line 581 "go.y"
    {
		(yyval.list) = variter((yyvsp[(1) - (3)].list), nil, (yyvsp[(3) - (3)].list));
	}
    break;

  case 65:
#line 587 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (4)].list), (yyvsp[(2) - (4)].node), (yyvsp[(4) - (4)].list));
	}
    break;

  case 66:
#line 591 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (3)].list), N, (yyvsp[(3) - (3)].list));
	}
    break;

  case 68:
#line 598 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].node), nil);
	}
    break;

  case 69:
#line 602 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (1)].list), N, nil);
	}
    break;

  case 70:
#line 608 "go.y"
    {
		// different from dclname because the name
		// becomes visible right here, not at the end
		// of the declaration.
		(yyval.node) = typedcl0((yyvsp[(1) - (1)].sym));
	}
    break;

  case 71:
#line 617 "go.y"
    {
		(yyval.node) = typedcl1((yyvsp[(1) - (2)].node), (yyvsp[(2) - (2)].node), 1);
	}
    break;

  case 72:
#line 623 "go.y"
    {
		(yyval.node) = (yyvsp[(1) - (1)].node);

		// These nodes do not carry line numbers.
		// Since a bare name used as an expression is an error,
		// introduce a wrapper node to give the correct line.
		switch((yyval.node)->op) {
		case ONAME:
		case ONONAME:
		case OTYPE:
		case OPACK:
		case OLITERAL:
			(yyval.node) = nod(OPAREN, (yyval.node), N);
			(yyval.node)->implicit = 1;
			break;
		}
	}
    break;

  case 73:
#line 641 "go.y"
    {
		(yyval.node) = nod(OASOP, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
		(yyval.node)->etype = (yyvsp[(2) - (3)].i);			// rathole to pass opcode
	}
    break;

  case 74:
#line 646 "go.y"
    {
		if((yyvsp[(1) - (3)].list)->next == nil && (yyvsp[(3) - (3)].list)->next == nil) {
			// simple
			(yyval.node) = nod(OAS, (yyvsp[(1) - (3)].list)->n, (yyvsp[(3) - (3)].list)->n);
			break;
		}
		// multiple
		(yyval.node) = nod(OAS2, N, N);
		(yyval.node)->list = (yyvsp[(1) - (3)].list);
		(yyval.node)->rlist = (yyvsp[(3) - (3)].list);
	}
    break;

  case 75:
#line 658 "go.y"
    {
		if((yyvsp[(3) - (3)].list)->n->op == OTYPESW) {
			(yyval.node) = nod(OTYPESW, N, (yyvsp[(3) - (3)].list)->n->right);
			if((yyvsp[(3) - (3)].list)->next != nil)
				yyerror("expr.(type) must be alone in list");
			if((yyvsp[(1) - (3)].list)->next != nil)
				yyerror("argument count mismatch: %d = %d", count((yyvsp[(1) - (3)].list)), 1);
			else if(((yyvsp[(1) - (3)].list)->n->op != ONAME && (yyvsp[(1) - (3)].list)->n->op != OTYPE && (yyvsp[(1) - (3)].list)->n->op != ONONAME) || isblank((yyvsp[(1) - (3)].list)->n))
				yyerror("invalid variable name %N in type switch", (yyvsp[(1) - (3)].list)->n);
			else
				(yyval.node)->left = dclname((yyvsp[(1) - (3)].list)->n->sym);  // it's a colas, so must not re-use an oldname.
			break;
		}
		(yyval.node) = colas((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list), (yyvsp[(2) - (3)].i));
	}
    break;

  case 76:
#line 674 "go.y"
    {
		(yyval.node) = nod(OASOP, (yyvsp[(1) - (2)].node), nodintconst(1));
		(yyval.node)->implicit = 1;
		(yyval.node)->etype = OADD;
	}
    break;

  case 77:
#line 680 "go.y"
    {
		(yyval.node) = nod(OASOP, (yyvsp[(1) - (2)].node), nodintconst(1));
		(yyval.node)->implicit = 1;
		(yyval.node)->etype = OSUB;
	}
    break;

  case 78:
#line 688 "go.y"
    {
		Node *n, *nn;

		// will be converted to OCASE
		// right will point to next case
		// done in casebody()
		markdcl();
		(yyval.node) = nod(OXCASE, N, N);
		(yyval.node)->list = (yyvsp[(2) - (3)].list);
		if(typesw != N && typesw->right != N && (n=typesw->right->left) != N) {
			// type switch - declare variable
			nn = newname(n->sym);
			declare(nn, dclcontext);
			(yyval.node)->nname = nn;

			// keep track of the instances for reporting unused
			nn->defn = typesw->right;
		}
	}
    break;

  case 79:
#line 708 "go.y"
    {
		Node *n;

		// will be converted to OCASE
		// right will point to next case
		// done in casebody()
		markdcl();
		(yyval.node) = nod(OXCASE, N, N);
		if((yyvsp[(2) - (5)].list)->next == nil)
			n = nod(OAS, (yyvsp[(2) - (5)].list)->n, (yyvsp[(4) - (5)].node));
		else {
			n = nod(OAS2, N, N);
			n->list = (yyvsp[(2) - (5)].list);
			n->rlist = list1((yyvsp[(4) - (5)].node));
		}
		(yyval.node)->list = list1(n);
	}
    break;

  case 80:
#line 726 "go.y"
    {
		// will be converted to OCASE
		// right will point to next case
		// done in casebody()
		markdcl();
		(yyval.node) = nod(OXCASE, N, N);
		(yyval.node)->list = list1(colas((yyvsp[(2) - (5)].list), list1((yyvsp[(4) - (5)].node)), (yyvsp[(3) - (5)].i)));
	}
    break;

  case 81:
#line 735 "go.y"
    {
		Node *n, *nn;

		markdcl();
		(yyval.node) = nod(OXCASE, N, N);
		if(typesw != N && typesw->right != N && (n=typesw->right->left) != N) {
			// type switch - declare variable
			nn = newname(n->sym);
			declare(nn, dclcontext);
			(yyval.node)->nname = nn;

			// keep track of the instances for reporting unused
			nn->defn = typesw->right;
		}
	}
    break;

  case 82:
#line 753 "go.y"
    {
		markdcl();
	}
    break;

  case 83:
#line 757 "go.y"
    {
		if((yyvsp[(3) - (4)].list) == nil)
			(yyval.node) = nod(OEMPTY, N, N);
		else
			(yyval.node) = liststmt((yyvsp[(3) - (4)].list));
		popdcl();
	}
    break;

  case 84:
#line 767 "go.y"
    {
		// If the last token read by the lexer was consumed
		// as part of the case, clear it (parser has cleared yychar).
		// If the last token read by the lexer was the lookahead
		// leave it alone (parser has it cached in yychar).
		// This is so that the stmt_list action doesn't look at
		// the case tokens if the stmt_list is empty.
		yylast = yychar;
		(yyvsp[(1) - (1)].node)->xoffset = block;
	}
    break;

  case 85:
#line 778 "go.y"
    {
		int last;

		// This is the only place in the language where a statement
		// list is not allowed to drop the final semicolon, because
		// it's the only place where a statement list is not followed 
		// by a closing brace.  Handle the error for pedantry.

		// Find the final token of the statement list.
		// yylast is lookahead; yyprev is last of stmt_list
		last = yyprev;

		if(last > 0 && last != ';' && yychar != '}')
			yyerror("missing statement after label");
		(yyval.node) = (yyvsp[(1) - (3)].node);
		(yyval.node)->nbody = (yyvsp[(3) - (3)].list);
		popdcl();
	}
    break;

  case 86:
#line 798 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 87:
#line 802 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].node));
	}
    break;

  case 88:
#line 808 "go.y"
    {
		markdcl();
	}
    break;

  case 89:
#line 812 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (4)].list);
		popdcl();
	}
    break;

  case 90:
#line 819 "go.y"
    {
		(yyval.node) = nod(ORANGE, N, (yyvsp[(4) - (4)].node));
		(yyval.node)->list = (yyvsp[(1) - (4)].list);
		(yyval.node)->etype = 0;	// := flag
	}
    break;

  case 91:
#line 825 "go.y"
    {
		(yyval.node) = nod(ORANGE, N, (yyvsp[(4) - (4)].node));
		(yyval.node)->list = (yyvsp[(1) - (4)].list);
		(yyval.node)->colas = 1;
		colasdefn((yyvsp[(1) - (4)].list), (yyval.node));
	}
    break;

  case 92:
#line 832 "go.y"
    {
		(yyval.node) = nod(ORANGE, N, (yyvsp[(2) - (2)].node));
		(yyval.node)->etype = 0; // := flag
	}
    break;

  case 93:
#line 839 "go.y"
    {
		// init ; test ; incr
		if((yyvsp[(5) - (5)].node) != N && (yyvsp[(5) - (5)].node)->colas != 0)
			yyerror("cannot declare in the for-increment");
		(yyval.node) = nod(OFOR, N, N);
		if((yyvsp[(1) - (5)].node) != N)
			(yyval.node)->ninit = list1((yyvsp[(1) - (5)].node));
		(yyval.node)->ntest = (yyvsp[(3) - (5)].node);
		(yyval.node)->nincr = (yyvsp[(5) - (5)].node);
	}
    break;

  case 94:
#line 850 "go.y"
    {
		// normal test
		(yyval.node) = nod(OFOR, N, N);
		(yyval.node)->ntest = (yyvsp[(1) - (1)].node);
	}
    break;

  case 96:
#line 859 "go.y"
    {
		(yyval.node) = (yyvsp[(1) - (2)].node);
		(yyval.node)->nbody = concat((yyval.node)->nbody, (yyvsp[(2) - (2)].list));
	}
    break;

  case 97:
#line 866 "go.y"
    {
		markdcl();
	}
    break;

  case 98:
#line 870 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (3)].node);
		popdcl();
	}
    break;

  case 99:
#line 877 "go.y"
    {
		// test
		(yyval.node) = nod(OIF, N, N);
		(yyval.node)->ntest = (yyvsp[(1) - (1)].node);
	}
    break;

  case 100:
#line 883 "go.y"
    {
		// init ; test
		(yyval.node) = nod(OIF, N, N);
		if((yyvsp[(1) - (3)].node) != N)
			(yyval.node)->ninit = list1((yyvsp[(1) - (3)].node));
		(yyval.node)->ntest = (yyvsp[(3) - (3)].node);
	}
    break;

  case 101:
#line 894 "go.y"
    {
		markdcl();
	}
    break;

  case 102:
#line 898 "go.y"
    {
		if((yyvsp[(3) - (3)].node)->ntest == N)
			yyerror("missing condition in if statement");
	}
    break;

  case 103:
#line 903 "go.y"
    {
		(yyvsp[(3) - (5)].node)->nbody = (yyvsp[(5) - (5)].list);
	}
    break;

  case 104:
#line 907 "go.y"
    {
		Node *n;
		NodeList *nn;

		(yyval.node) = (yyvsp[(3) - (8)].node);
		n = (yyvsp[(3) - (8)].node);
		popdcl();
		for(nn = concat((yyvsp[(7) - (8)].list), (yyvsp[(8) - (8)].list)); nn; nn = nn->next) {
			if(nn->n->op == OIF)
				popdcl();
			n->nelse = list1(nn->n);
			n = nn->n;
		}
	}
    break;

  case 105:
#line 924 "go.y"
    {
		markdcl();
	}
    break;

  case 106:
#line 928 "go.y"
    {
		if((yyvsp[(4) - (5)].node)->ntest == N)
			yyerror("missing condition in if statement");
		(yyvsp[(4) - (5)].node)->nbody = (yyvsp[(5) - (5)].list);
		(yyval.list) = list1((yyvsp[(4) - (5)].node));
	}
    break;

  case 107:
#line 936 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 108:
#line 940 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].list));
	}
    break;

  case 109:
#line 945 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 110:
#line 949 "go.y"
    {
		NodeList *node;
		
		node = mal(sizeof *node);
		node->n = (yyvsp[(2) - (2)].node);
		node->end = node;
		(yyval.list) = node;
	}
    break;

  case 111:
#line 960 "go.y"
    {
		markdcl();
	}
    break;

  case 112:
#line 964 "go.y"
    {
		Node *n;
		n = (yyvsp[(3) - (3)].node)->ntest;
		if(n != N && n->op != OTYPESW)
			n = N;
		typesw = nod(OXXX, typesw, n);
	}
    break;

  case 113:
#line 972 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (7)].node);
		(yyval.node)->op = OSWITCH;
		(yyval.node)->list = (yyvsp[(6) - (7)].list);
		typesw = typesw->left;
		popdcl();
	}
    break;

  case 114:
#line 982 "go.y"
    {
		typesw = nod(OXXX, typesw, N);
	}
    break;

  case 115:
#line 986 "go.y"
    {
		(yyval.node) = nod(OSELECT, N, N);
		(yyval.node)->lineno = typesw->lineno;
		(yyval.node)->list = (yyvsp[(4) - (5)].list);
		typesw = typesw->left;
	}
    break;

  case 117:
#line 999 "go.y"
    {
		(yyval.node) = nod(OOROR, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 118:
#line 1003 "go.y"
    {
		(yyval.node) = nod(OANDAND, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 119:
#line 1007 "go.y"
    {
		(yyval.node) = nod(OEQ, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 120:
#line 1011 "go.y"
    {
		(yyval.node) = nod(ONE, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 121:
#line 1015 "go.y"
    {
		(yyval.node) = nod(OLT, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 122:
#line 1019 "go.y"
    {
		(yyval.node) = nod(OLE, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 123:
#line 1023 "go.y"
    {
		(yyval.node) = nod(OGE, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 124:
#line 1027 "go.y"
    {
		(yyval.node) = nod(OGT, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 125:
#line 1031 "go.y"
    {
		(yyval.node) = nod(OADD, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 126:
#line 1035 "go.y"
    {
		(yyval.node) = nod(OSUB, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 127:
#line 1039 "go.y"
    {
		(yyval.node) = nod(OOR, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 128:
#line 1043 "go.y"
    {
		(yyval.node) = nod(OXOR, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 129:
#line 1047 "go.y"
    {
		(yyval.node) = nod(OMUL, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 130:
#line 1051 "go.y"
    {
		(yyval.node) = nod(ODIV, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 131:
#line 1055 "go.y"
    {
		(yyval.node) = nod(OMOD, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 132:
#line 1059 "go.y"
    {
		(yyval.node) = nod(OAND, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 133:
#line 1063 "go.y"
    {
		(yyval.node) = nod(OANDNOT, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 134:
#line 1067 "go.y"
    {
		(yyval.node) = nod(OLSH, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 135:
#line 1071 "go.y"
    {
		(yyval.node) = nod(ORSH, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 136:
#line 1076 "go.y"
    {
		(yyval.node) = nod(OSEND, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 138:
#line 1083 "go.y"
    {
		(yyval.node) = nod(OIND, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 139:
#line 1087 "go.y"
    {
		if((yyvsp[(2) - (2)].node)->op == OCOMPLIT) {
			// Special case for &T{...}: turn into (*T){...}.
			(yyval.node) = (yyvsp[(2) - (2)].node);
			(yyval.node)->right = nod(OIND, (yyval.node)->right, N);
			(yyval.node)->right->implicit = 1;
		} else {
			(yyval.node) = nod(OADDR, (yyvsp[(2) - (2)].node), N);
		}
	}
    break;

  case 140:
#line 1098 "go.y"
    {
		(yyval.node) = nod(OPLUS, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 141:
#line 1102 "go.y"
    {
		(yyval.node) = nod(OMINUS, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 142:
#line 1106 "go.y"
    {
		(yyval.node) = nod(ONOT, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 143:
#line 1110 "go.y"
    {
		yyerror("the bitwise complement operator is ^");
		(yyval.node) = nod(OCOM, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 144:
#line 1115 "go.y"
    {
		(yyval.node) = nod(OCOM, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 145:
#line 1119 "go.y"
    {
		(yyval.node) = nod(ORECV, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 146:
#line 1129 "go.y"
    {
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (3)].node), N);
	}
    break;

  case 147:
#line 1133 "go.y"
    {
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (5)].node), N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
	}
    break;

  case 148:
#line 1138 "go.y"
    {
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (6)].node), N);
		(yyval.node)->list = (yyvsp[(3) - (6)].list);
		(yyval.node)->isddd = 1;
	}
    break;

  case 149:
#line 1146 "go.y"
    {
		(yyval.node) = nodlit((yyvsp[(1) - (1)].val));
	}
    break;

  case 151:
#line 1151 "go.y"
    {
		if((yyvsp[(1) - (3)].node)->op == OPACK) {
			Sym *s;
			s = restrictlookup((yyvsp[(3) - (3)].sym)->name, (yyvsp[(1) - (3)].node)->pkg);
			(yyvsp[(1) - (3)].node)->used = 1;
			(yyval.node) = oldname(s);
			break;
		}
		(yyval.node) = nod(OXDOT, (yyvsp[(1) - (3)].node), newname((yyvsp[(3) - (3)].sym)));
	}
    break;

  case 152:
#line 1162 "go.y"
    {
		(yyval.node) = nod(ODOTTYPE, (yyvsp[(1) - (5)].node), (yyvsp[(4) - (5)].node));
	}
    break;

  case 153:
#line 1166 "go.y"
    {
		(yyval.node) = nod(OTYPESW, N, (yyvsp[(1) - (5)].node));
	}
    break;

  case 154:
#line 1170 "go.y"
    {
		(yyval.node) = nod(OINDEX, (yyvsp[(1) - (4)].node), (yyvsp[(3) - (4)].node));
	}
    break;

  case 155:
#line 1174 "go.y"
    {
		(yyval.node) = nod(OSLICE, (yyvsp[(1) - (6)].node), nod(OKEY, (yyvsp[(3) - (6)].node), (yyvsp[(5) - (6)].node)));
	}
    break;

  case 156:
#line 1178 "go.y"
    {
		if((yyvsp[(5) - (8)].node) == N)
			yyerror("middle index required in 3-index slice");
		if((yyvsp[(7) - (8)].node) == N)
			yyerror("final index required in 3-index slice");
		(yyval.node) = nod(OSLICE3, (yyvsp[(1) - (8)].node), nod(OKEY, (yyvsp[(3) - (8)].node), nod(OKEY, (yyvsp[(5) - (8)].node), (yyvsp[(7) - (8)].node))));
	}
    break;

  case 158:
#line 1187 "go.y"
    {
		// conversion
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (5)].node), N);
		(yyval.node)->list = list1((yyvsp[(3) - (5)].node));
	}
    break;

  case 159:
#line 1193 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (5)].node);
		(yyval.node)->right = (yyvsp[(1) - (5)].node);
		(yyval.node)->list = (yyvsp[(4) - (5)].list);
		fixlbrace((yyvsp[(2) - (5)].i));
	}
    break;

  case 160:
#line 1200 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (5)].node);
		(yyval.node)->right = (yyvsp[(1) - (5)].node);
		(yyval.node)->list = (yyvsp[(4) - (5)].list);
	}
    break;

  case 161:
#line 1206 "go.y"
    {
		yyerror("cannot parenthesize type in composite literal");
		(yyval.node) = (yyvsp[(5) - (7)].node);
		(yyval.node)->right = (yyvsp[(2) - (7)].node);
		(yyval.node)->list = (yyvsp[(6) - (7)].list);
	}
    break;

  case 163:
#line 1215 "go.y"
    {
		// composite expression.
		// make node early so we get the right line number.
		(yyval.node) = nod(OCOMPLIT, N, N);
	}
    break;

  case 164:
#line 1223 "go.y"
    {
		(yyval.node) = nod(OKEY, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 165:
#line 1229 "go.y"
    {
		// These nodes do not carry line numbers.
		// Since a composite literal commonly spans several lines,
		// the line number on errors may be misleading.
		// Introduce a wrapper node to give the correct line.
		(yyval.node) = (yyvsp[(1) - (1)].node);
		switch((yyval.node)->op) {
		case ONAME:
		case ONONAME:
		case OTYPE:
		case OPACK:
		case OLITERAL:
			(yyval.node) = nod(OPAREN, (yyval.node), N);
			(yyval.node)->implicit = 1;
		}
	}
    break;

  case 166:
#line 1246 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (4)].node);
		(yyval.node)->list = (yyvsp[(3) - (4)].list);
	}
    break;

  case 168:
#line 1254 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (4)].node);
		(yyval.node)->list = (yyvsp[(3) - (4)].list);
	}
    break;

  case 170:
#line 1262 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
		
		// Need to know on lhs of := whether there are ( ).
		// Don't bother with the OPAREN in other cases:
		// it's just a waste of memory and time.
		switch((yyval.node)->op) {
		case ONAME:
		case ONONAME:
		case OPACK:
		case OTYPE:
		case OLITERAL:
		case OTYPESW:
			(yyval.node) = nod(OPAREN, (yyval.node), N);
		}
	}
    break;

  case 174:
#line 1288 "go.y"
    {
		(yyval.i) = LBODY;
	}
    break;

  case 175:
#line 1292 "go.y"
    {
		(yyval.i) = '{';
	}
    break;

  case 176:
#line 1303 "go.y"
    {
		if((yyvsp[(1) - (1)].sym) == S)
			(yyval.node) = N;
		else
			(yyval.node) = newname((yyvsp[(1) - (1)].sym));
	}
    break;

  case 177:
#line 1312 "go.y"
    {
		(yyval.node) = dclname((yyvsp[(1) - (1)].sym));
	}
    break;

  case 178:
#line 1317 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 180:
#line 1324 "go.y"
    {
		(yyval.sym) = (yyvsp[(1) - (1)].sym);
		// during imports, unqualified non-exported identifiers are from builtinpkg
		if(importpkg != nil && !exportname((yyvsp[(1) - (1)].sym)->name))
			(yyval.sym) = pkglookup((yyvsp[(1) - (1)].sym)->name, builtinpkg);
	}
    break;

  case 182:
#line 1332 "go.y"
    {
		(yyval.sym) = S;
	}
    break;

  case 183:
#line 1338 "go.y"
    {
		Pkg *p;

		if((yyvsp[(2) - (4)].val).u.sval->len == 0)
			p = importpkg;
		else {
			if(isbadimport((yyvsp[(2) - (4)].val).u.sval))
				errorexit();
			p = mkpkg((yyvsp[(2) - (4)].val).u.sval);
		}
		(yyval.sym) = pkglookup((yyvsp[(4) - (4)].sym)->name, p);
	}
    break;

  case 184:
#line 1351 "go.y"
    {
		Pkg *p;

		if((yyvsp[(2) - (4)].val).u.sval->len == 0)
			p = importpkg;
		else {
			if(isbadimport((yyvsp[(2) - (4)].val).u.sval))
				errorexit();
			p = mkpkg((yyvsp[(2) - (4)].val).u.sval);
		}
		(yyval.sym) = pkglookup("?", p);
	}
    break;

  case 185:
#line 1366 "go.y"
    {
		(yyval.node) = oldname((yyvsp[(1) - (1)].sym));
		if((yyval.node)->pack != N)
			(yyval.node)->pack->used = 1;
	}
    break;

  case 187:
#line 1386 "go.y"
    {
		yyerror("final argument in variadic function missing type");
		(yyval.node) = nod(ODDD, typenod(typ(TINTER)), N);
	}
    break;

  case 188:
#line 1391 "go.y"
    {
		(yyval.node) = nod(ODDD, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 194:
#line 1402 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
	}
    break;

  case 198:
#line 1411 "go.y"
    {
		(yyval.node) = nod(OIND, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 203:
#line 1421 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
	}
    break;

  case 213:
#line 1442 "go.y"
    {
		if((yyvsp[(1) - (3)].node)->op == OPACK) {
			Sym *s;
			s = restrictlookup((yyvsp[(3) - (3)].sym)->name, (yyvsp[(1) - (3)].node)->pkg);
			(yyvsp[(1) - (3)].node)->used = 1;
			(yyval.node) = oldname(s);
			break;
		}
		(yyval.node) = nod(OXDOT, (yyvsp[(1) - (3)].node), newname((yyvsp[(3) - (3)].sym)));
	}
    break;

  case 214:
#line 1455 "go.y"
    {
		(yyval.node) = nod(OTARRAY, (yyvsp[(2) - (4)].node), (yyvsp[(4) - (4)].node));
	}
    break;

  case 215:
#line 1459 "go.y"
    {
		// array literal of nelem
		(yyval.node) = nod(OTARRAY, nod(ODDD, N, N), (yyvsp[(4) - (4)].node));
	}
    break;

  case 216:
#line 1464 "go.y"
    {
		(yyval.node) = nod(OTCHAN, (yyvsp[(2) - (2)].node), N);
		(yyval.node)->etype = Cboth;
	}
    break;

  case 217:
#line 1469 "go.y"
    {
		(yyval.node) = nod(OTCHAN, (yyvsp[(3) - (3)].node), N);
		(yyval.node)->etype = Csend;
	}
    break;

  case 218:
#line 1474 "go.y"
    {
		(yyval.node) = nod(OTMAP, (yyvsp[(3) - (5)].node), (yyvsp[(5) - (5)].node));
	}
    break;

  case 221:
#line 1482 "go.y"
    {
		(yyval.node) = nod(OIND, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 222:
#line 1488 "go.y"
    {
		(yyval.node) = nod(OTCHAN, (yyvsp[(3) - (3)].node), N);
		(yyval.node)->etype = Crecv;
	}
    break;

  case 223:
#line 1495 "go.y"
    {
		(yyval.node) = nod(OTSTRUCT, N, N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
		fixlbrace((yyvsp[(2) - (5)].i));
	}
    break;

  case 224:
#line 1501 "go.y"
    {
		(yyval.node) = nod(OTSTRUCT, N, N);
		fixlbrace((yyvsp[(2) - (3)].i));
	}
    break;

  case 225:
#line 1508 "go.y"
    {
		(yyval.node) = nod(OTINTER, N, N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
		fixlbrace((yyvsp[(2) - (5)].i));
	}
    break;

  case 226:
#line 1514 "go.y"
    {
		(yyval.node) = nod(OTINTER, N, N);
		fixlbrace((yyvsp[(2) - (3)].i));
	}
    break;

  case 227:
#line 1525 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
		if((yyval.node) == N)
			break;
		if(noescape && (yyvsp[(3) - (3)].list) != nil)
			yyerror("can only use //go:noescape with external func implementations");
		(yyval.node)->nbody = (yyvsp[(3) - (3)].list);
		(yyval.node)->endlineno = lineno;
		(yyval.node)->noescape = noescape;
		(yyval.node)->nosplit = nosplit;
		(yyval.node)->nowritebarrier = nowritebarrier;
		funcbody((yyval.node));
	}
    break;

  case 228:
#line 1541 "go.y"
    {
		Node *t;

		(yyval.node) = N;
		(yyvsp[(3) - (5)].list) = checkarglist((yyvsp[(3) - (5)].list), 1);

		if(strcmp((yyvsp[(1) - (5)].sym)->name, "init") == 0) {
			(yyvsp[(1) - (5)].sym) = renameinit();
			if((yyvsp[(3) - (5)].list) != nil || (yyvsp[(5) - (5)].list) != nil)
				yyerror("func init must have no arguments and no return values");
		}
		if(strcmp(localpkg->name, "main") == 0 && strcmp((yyvsp[(1) - (5)].sym)->name, "main") == 0) {
			if((yyvsp[(3) - (5)].list) != nil || (yyvsp[(5) - (5)].list) != nil)
				yyerror("func main must have no arguments and no return values");
		}

		t = nod(OTFUNC, N, N);
		t->list = (yyvsp[(3) - (5)].list);
		t->rlist = (yyvsp[(5) - (5)].list);

		(yyval.node) = nod(ODCLFUNC, N, N);
		(yyval.node)->nname = newname((yyvsp[(1) - (5)].sym));
		(yyval.node)->nname->defn = (yyval.node);
		(yyval.node)->nname->ntype = t;		// TODO: check if nname already has an ntype
		declare((yyval.node)->nname, PFUNC);

		funchdr((yyval.node));
	}
    break;

  case 229:
#line 1570 "go.y"
    {
		Node *rcvr, *t;

		(yyval.node) = N;
		(yyvsp[(2) - (8)].list) = checkarglist((yyvsp[(2) - (8)].list), 0);
		(yyvsp[(6) - (8)].list) = checkarglist((yyvsp[(6) - (8)].list), 1);

		if((yyvsp[(2) - (8)].list) == nil) {
			yyerror("method has no receiver");
			break;
		}
		if((yyvsp[(2) - (8)].list)->next != nil) {
			yyerror("method has multiple receivers");
			break;
		}
		rcvr = (yyvsp[(2) - (8)].list)->n;
		if(rcvr->op != ODCLFIELD) {
			yyerror("bad receiver in method");
			break;
		}

		t = nod(OTFUNC, rcvr, N);
		t->list = (yyvsp[(6) - (8)].list);
		t->rlist = (yyvsp[(8) - (8)].list);

		(yyval.node) = nod(ODCLFUNC, N, N);
		(yyval.node)->shortname = newname((yyvsp[(4) - (8)].sym));
		(yyval.node)->nname = methodname1((yyval.node)->shortname, rcvr->right);
		(yyval.node)->nname->defn = (yyval.node);
		(yyval.node)->nname->ntype = t;
		(yyval.node)->nname->nointerface = nointerface;
		declare((yyval.node)->nname, PFUNC);

		funchdr((yyval.node));
	}
    break;

  case 230:
#line 1608 "go.y"
    {
		Sym *s;
		Type *t;

		(yyval.node) = N;

		s = (yyvsp[(1) - (5)].sym);
		t = functype(N, (yyvsp[(3) - (5)].list), (yyvsp[(5) - (5)].list));

		importsym(s, ONAME);
		if(s->def != N && s->def->op == ONAME) {
			if(eqtype(t, s->def->type)) {
				dclcontext = PDISCARD;  // since we skip funchdr below
				break;
			}
			yyerror("inconsistent definition for func %S during import\n\t%T\n\t%T", s, s->def->type, t);
		}

		(yyval.node) = newname(s);
		(yyval.node)->type = t;
		declare((yyval.node), PFUNC);

		funchdr((yyval.node));
	}
    break;

  case 231:
#line 1633 "go.y"
    {
		(yyval.node) = methodname1(newname((yyvsp[(4) - (8)].sym)), (yyvsp[(2) - (8)].list)->n->right); 
		(yyval.node)->type = functype((yyvsp[(2) - (8)].list)->n, (yyvsp[(6) - (8)].list), (yyvsp[(8) - (8)].list));

		checkwidth((yyval.node)->type);
		addmethod((yyvsp[(4) - (8)].sym), (yyval.node)->type, 0, nointerface);
		nointerface = 0;
		funchdr((yyval.node));
		
		// inl.c's inlnode in on a dotmeth node expects to find the inlineable body as
		// (dotmeth's type)->nname->inl, and dotmeth's type has been pulled
		// out by typecheck's lookdot as this $$->ttype.  So by providing
		// this back link here we avoid special casing there.
		(yyval.node)->type->nname = (yyval.node);
	}
    break;

  case 232:
#line 1651 "go.y"
    {
		(yyvsp[(3) - (5)].list) = checkarglist((yyvsp[(3) - (5)].list), 1);
		(yyval.node) = nod(OTFUNC, N, N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
		(yyval.node)->rlist = (yyvsp[(5) - (5)].list);
	}
    break;

  case 233:
#line 1659 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 234:
#line 1663 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (3)].list);
		if((yyval.list) == nil)
			(yyval.list) = list1(nod(OEMPTY, N, N));
	}
    break;

  case 235:
#line 1671 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 236:
#line 1675 "go.y"
    {
		(yyval.list) = list1(nod(ODCLFIELD, N, (yyvsp[(1) - (1)].node)));
	}
    break;

  case 237:
#line 1679 "go.y"
    {
		(yyvsp[(2) - (3)].list) = checkarglist((yyvsp[(2) - (3)].list), 0);
		(yyval.list) = (yyvsp[(2) - (3)].list);
	}
    break;

  case 238:
#line 1686 "go.y"
    {
		closurehdr((yyvsp[(1) - (1)].node));
	}
    break;

  case 239:
#line 1692 "go.y"
    {
		(yyval.node) = closurebody((yyvsp[(3) - (4)].list));
		fixlbrace((yyvsp[(2) - (4)].i));
	}
    break;

  case 240:
#line 1697 "go.y"
    {
		(yyval.node) = closurebody(nil);
	}
    break;

  case 241:
#line 1708 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 242:
#line 1712 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(2) - (3)].list));
		if(nsyntaxerrors == 0)
			testdclstack();
		nointerface = 0;
		noescape = 0;
		nosplit = 0;
		nowritebarrier = 0;
	}
    break;

  case 244:
#line 1725 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list));
	}
    break;

  case 246:
#line 1732 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list));
	}
    break;

  case 247:
#line 1738 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 248:
#line 1742 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 250:
#line 1749 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list));
	}
    break;

  case 251:
#line 1755 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 252:
#line 1759 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 253:
#line 1765 "go.y"
    {
		NodeList *l;

		Node *n;
		l = (yyvsp[(1) - (3)].list);
		if(l == nil) {
			// ? symbol, during import (list1(N) == nil)
			n = (yyvsp[(2) - (3)].node);
			if(n->op == OIND)
				n = n->left;
			n = embedded(n->sym, importpkg);
			n->right = (yyvsp[(2) - (3)].node);
			n->val = (yyvsp[(3) - (3)].val);
			(yyval.list) = list1(n);
			break;
		}

		for(l=(yyvsp[(1) - (3)].list); l; l=l->next) {
			l->n = nod(ODCLFIELD, l->n, (yyvsp[(2) - (3)].node));
			l->n->val = (yyvsp[(3) - (3)].val);
		}
	}
    break;

  case 254:
#line 1788 "go.y"
    {
		(yyvsp[(1) - (2)].node)->val = (yyvsp[(2) - (2)].val);
		(yyval.list) = list1((yyvsp[(1) - (2)].node));
	}
    break;

  case 255:
#line 1793 "go.y"
    {
		(yyvsp[(2) - (4)].node)->val = (yyvsp[(4) - (4)].val);
		(yyval.list) = list1((yyvsp[(2) - (4)].node));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 256:
#line 1799 "go.y"
    {
		(yyvsp[(2) - (3)].node)->right = nod(OIND, (yyvsp[(2) - (3)].node)->right, N);
		(yyvsp[(2) - (3)].node)->val = (yyvsp[(3) - (3)].val);
		(yyval.list) = list1((yyvsp[(2) - (3)].node));
	}
    break;

  case 257:
#line 1805 "go.y"
    {
		(yyvsp[(3) - (5)].node)->right = nod(OIND, (yyvsp[(3) - (5)].node)->right, N);
		(yyvsp[(3) - (5)].node)->val = (yyvsp[(5) - (5)].val);
		(yyval.list) = list1((yyvsp[(3) - (5)].node));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 258:
#line 1812 "go.y"
    {
		(yyvsp[(3) - (5)].node)->right = nod(OIND, (yyvsp[(3) - (5)].node)->right, N);
		(yyvsp[(3) - (5)].node)->val = (yyvsp[(5) - (5)].val);
		(yyval.list) = list1((yyvsp[(3) - (5)].node));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 259:
#line 1821 "go.y"
    {
		Node *n;

		(yyval.sym) = (yyvsp[(1) - (1)].sym);
		n = oldname((yyvsp[(1) - (1)].sym));
		if(n->pack != N)
			n->pack->used = 1;
	}
    break;

  case 260:
#line 1830 "go.y"
    {
		Pkg *pkg;

		if((yyvsp[(1) - (3)].sym)->def == N || (yyvsp[(1) - (3)].sym)->def->op != OPACK) {
			yyerror("%S is not a package", (yyvsp[(1) - (3)].sym));
			pkg = localpkg;
		} else {
			(yyvsp[(1) - (3)].sym)->def->used = 1;
			pkg = (yyvsp[(1) - (3)].sym)->def->pkg;
		}
		(yyval.sym) = restrictlookup((yyvsp[(3) - (3)].sym)->name, pkg);
	}
    break;

  case 261:
#line 1845 "go.y"
    {
		(yyval.node) = embedded((yyvsp[(1) - (1)].sym), localpkg);
	}
    break;

  case 262:
#line 1851 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, (yyvsp[(1) - (2)].node), (yyvsp[(2) - (2)].node));
		ifacedcl((yyval.node));
	}
    break;

  case 263:
#line 1856 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, oldname((yyvsp[(1) - (1)].sym)));
	}
    break;

  case 264:
#line 1860 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, oldname((yyvsp[(2) - (3)].sym)));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 265:
#line 1867 "go.y"
    {
		// without func keyword
		(yyvsp[(2) - (4)].list) = checkarglist((yyvsp[(2) - (4)].list), 1);
		(yyval.node) = nod(OTFUNC, fakethis(), N);
		(yyval.node)->list = (yyvsp[(2) - (4)].list);
		(yyval.node)->rlist = (yyvsp[(4) - (4)].list);
	}
    break;

  case 267:
#line 1881 "go.y"
    {
		(yyval.node) = nod(ONONAME, N, N);
		(yyval.node)->sym = (yyvsp[(1) - (2)].sym);
		(yyval.node) = nod(OKEY, (yyval.node), (yyvsp[(2) - (2)].node));
	}
    break;

  case 268:
#line 1887 "go.y"
    {
		(yyval.node) = nod(ONONAME, N, N);
		(yyval.node)->sym = (yyvsp[(1) - (2)].sym);
		(yyval.node) = nod(OKEY, (yyval.node), (yyvsp[(2) - (2)].node));
	}
    break;

  case 270:
#line 1896 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 271:
#line 1900 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 272:
#line 1905 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 273:
#line 1909 "go.y"
    {
		(yyval.list) = (yyvsp[(1) - (2)].list);
	}
    break;

  case 274:
#line 1917 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 276:
#line 1922 "go.y"
    {
		(yyval.node) = liststmt((yyvsp[(1) - (1)].list));
	}
    break;

  case 278:
#line 1927 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 284:
#line 1938 "go.y"
    {
		(yyvsp[(1) - (2)].node) = nod(OLABEL, (yyvsp[(1) - (2)].node), N);
		(yyvsp[(1) - (2)].node)->sym = dclstack;  // context, for goto restrictions
	}
    break;

  case 285:
#line 1943 "go.y"
    {
		NodeList *l;

		(yyvsp[(1) - (4)].node)->defn = (yyvsp[(4) - (4)].node);
		l = list1((yyvsp[(1) - (4)].node));
		if((yyvsp[(4) - (4)].node))
			l = list(l, (yyvsp[(4) - (4)].node));
		(yyval.node) = liststmt(l);
	}
    break;

  case 286:
#line 1953 "go.y"
    {
		// will be converted to OFALL
		(yyval.node) = nod(OXFALL, N, N);
		(yyval.node)->xoffset = block;
	}
    break;

  case 287:
#line 1959 "go.y"
    {
		(yyval.node) = nod(OBREAK, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 288:
#line 1963 "go.y"
    {
		(yyval.node) = nod(OCONTINUE, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 289:
#line 1967 "go.y"
    {
		(yyval.node) = nod(OPROC, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 290:
#line 1971 "go.y"
    {
		(yyval.node) = nod(ODEFER, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 291:
#line 1975 "go.y"
    {
		(yyval.node) = nod(OGOTO, (yyvsp[(2) - (2)].node), N);
		(yyval.node)->sym = dclstack;  // context, for goto restrictions
	}
    break;

  case 292:
#line 1980 "go.y"
    {
		(yyval.node) = nod(ORETURN, N, N);
		(yyval.node)->list = (yyvsp[(2) - (2)].list);
		if((yyval.node)->list == nil && curfn != N) {
			NodeList *l;

			for(l=curfn->dcl; l; l=l->next) {
				if(l->n->class == PPARAM)
					continue;
				if(l->n->class != PPARAMOUT)
					break;
				if(l->n->sym->def != l->n)
					yyerror("%s is shadowed during return", l->n->sym->name);
			}
		}
	}
    break;

  case 293:
#line 1999 "go.y"
    {
		(yyval.list) = nil;
		if((yyvsp[(1) - (1)].node) != N)
			(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 294:
#line 2005 "go.y"
    {
		(yyval.list) = (yyvsp[(1) - (3)].list);
		if((yyvsp[(3) - (3)].node) != N)
			(yyval.list) = list((yyval.list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 295:
#line 2013 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 296:
#line 2017 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 297:
#line 2023 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 298:
#line 2027 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 299:
#line 2033 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 300:
#line 2037 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 301:
#line 2043 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 302:
#line 2047 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 303:
#line 2056 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 304:
#line 2060 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 305:
#line 2064 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 306:
#line 2068 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 307:
#line 2073 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 308:
#line 2077 "go.y"
    {
		(yyval.list) = (yyvsp[(1) - (2)].list);
	}
    break;

  case 313:
#line 2091 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 315:
#line 2097 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 317:
#line 2103 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 319:
#line 2109 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 321:
#line 2115 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 323:
#line 2121 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 325:
#line 2127 "go.y"
    {
		(yyval.val).ctype = CTxxx;
	}
    break;

  case 327:
#line 2137 "go.y"
    {
		importimport((yyvsp[(2) - (4)].sym), (yyvsp[(3) - (4)].val).u.sval);
	}
    break;

  case 328:
#line 2141 "go.y"
    {
		importvar((yyvsp[(2) - (4)].sym), (yyvsp[(3) - (4)].type));
	}
    break;

  case 329:
#line 2145 "go.y"
    {
		importconst((yyvsp[(2) - (5)].sym), types[TIDEAL], (yyvsp[(4) - (5)].node));
	}
    break;

  case 330:
#line 2149 "go.y"
    {
		importconst((yyvsp[(2) - (6)].sym), (yyvsp[(3) - (6)].type), (yyvsp[(5) - (6)].node));
	}
    break;

  case 331:
#line 2153 "go.y"
    {
		importtype((yyvsp[(2) - (4)].type), (yyvsp[(3) - (4)].type));
	}
    break;

  case 332:
#line 2157 "go.y"
    {
		if((yyvsp[(2) - (4)].node) == N) {
			dclcontext = PEXTERN;  // since we skip the funcbody below
			break;
		}

		(yyvsp[(2) - (4)].node)->inl = (yyvsp[(3) - (4)].list);

		funcbody((yyvsp[(2) - (4)].node));
		importlist = list(importlist, (yyvsp[(2) - (4)].node));

		if(debug['E']) {
			print("import [%Z] func %lN \n", importpkg->path, (yyvsp[(2) - (4)].node));
			if(debug['m'] > 2 && (yyvsp[(2) - (4)].node)->inl)
				print("inl body:%+H\n", (yyvsp[(2) - (4)].node)->inl);
		}
	}
    break;

  case 333:
#line 2177 "go.y"
    {
		(yyval.sym) = (yyvsp[(1) - (1)].sym);
		structpkg = (yyval.sym)->pkg;
	}
    break;

  case 334:
#line 2184 "go.y"
    {
		(yyval.type) = pkgtype((yyvsp[(1) - (1)].sym));
		importsym((yyvsp[(1) - (1)].sym), OTYPE);
	}
    break;

  case 340:
#line 2204 "go.y"
    {
		(yyval.type) = pkgtype((yyvsp[(1) - (1)].sym));
	}
    break;

  case 341:
#line 2208 "go.y"
    {
		// predefined name like uint8
		(yyvsp[(1) - (1)].sym) = pkglookup((yyvsp[(1) - (1)].sym)->name, builtinpkg);
		if((yyvsp[(1) - (1)].sym)->def == N || (yyvsp[(1) - (1)].sym)->def->op != OTYPE) {
			yyerror("%s is not a type", (yyvsp[(1) - (1)].sym)->name);
			(yyval.type) = T;
		} else
			(yyval.type) = (yyvsp[(1) - (1)].sym)->def->type;
	}
    break;

  case 342:
#line 2218 "go.y"
    {
		(yyval.type) = aindex(N, (yyvsp[(3) - (3)].type));
	}
    break;

  case 343:
#line 2222 "go.y"
    {
		(yyval.type) = aindex(nodlit((yyvsp[(2) - (4)].val)), (yyvsp[(4) - (4)].type));
	}
    break;

  case 344:
#line 2226 "go.y"
    {
		(yyval.type) = maptype((yyvsp[(3) - (5)].type), (yyvsp[(5) - (5)].type));
	}
    break;

  case 345:
#line 2230 "go.y"
    {
		(yyval.type) = tostruct((yyvsp[(3) - (4)].list));
	}
    break;

  case 346:
#line 2234 "go.y"
    {
		(yyval.type) = tointerface((yyvsp[(3) - (4)].list));
	}
    break;

  case 347:
#line 2238 "go.y"
    {
		(yyval.type) = ptrto((yyvsp[(2) - (2)].type));
	}
    break;

  case 348:
#line 2242 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(2) - (2)].type);
		(yyval.type)->chan = Cboth;
	}
    break;

  case 349:
#line 2248 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(3) - (4)].type);
		(yyval.type)->chan = Cboth;
	}
    break;

  case 350:
#line 2254 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(3) - (3)].type);
		(yyval.type)->chan = Csend;
	}
    break;

  case 351:
#line 2262 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(3) - (3)].type);
		(yyval.type)->chan = Crecv;
	}
    break;

  case 352:
#line 2270 "go.y"
    {
		(yyval.type) = functype(nil, (yyvsp[(3) - (5)].list), (yyvsp[(5) - (5)].list));
	}
    break;

  case 353:
#line 2276 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, typenod((yyvsp[(2) - (3)].type)));
		if((yyvsp[(1) - (3)].sym))
			(yyval.node)->left = newname((yyvsp[(1) - (3)].sym));
		(yyval.node)->val = (yyvsp[(3) - (3)].val);
	}
    break;

  case 354:
#line 2283 "go.y"
    {
		Type *t;
	
		t = typ(TARRAY);
		t->bound = -1;
		t->type = (yyvsp[(3) - (4)].type);

		(yyval.node) = nod(ODCLFIELD, N, typenod(t));
		if((yyvsp[(1) - (4)].sym))
			(yyval.node)->left = newname((yyvsp[(1) - (4)].sym));
		(yyval.node)->isddd = 1;
		(yyval.node)->val = (yyvsp[(4) - (4)].val);
	}
    break;

  case 355:
#line 2299 "go.y"
    {
		Sym *s;
		Pkg *p;

		if((yyvsp[(1) - (3)].sym) != S && strcmp((yyvsp[(1) - (3)].sym)->name, "?") != 0) {
			(yyval.node) = nod(ODCLFIELD, newname((yyvsp[(1) - (3)].sym)), typenod((yyvsp[(2) - (3)].type)));
			(yyval.node)->val = (yyvsp[(3) - (3)].val);
		} else {
			s = (yyvsp[(2) - (3)].type)->sym;
			if(s == S && isptr[(yyvsp[(2) - (3)].type)->etype])
				s = (yyvsp[(2) - (3)].type)->type->sym;
			p = importpkg;
			if((yyvsp[(1) - (3)].sym) != S)
				p = (yyvsp[(1) - (3)].sym)->pkg;
			(yyval.node) = embedded(s, p);
			(yyval.node)->right = typenod((yyvsp[(2) - (3)].type));
			(yyval.node)->val = (yyvsp[(3) - (3)].val);
		}
	}
    break;

  case 356:
#line 2321 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, newname((yyvsp[(1) - (5)].sym)), typenod(functype(fakethis(), (yyvsp[(3) - (5)].list), (yyvsp[(5) - (5)].list))));
	}
    break;

  case 357:
#line 2325 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, typenod((yyvsp[(1) - (1)].type)));
	}
    break;

  case 358:
#line 2330 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 360:
#line 2337 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (3)].list);
	}
    break;

  case 361:
#line 2341 "go.y"
    {
		(yyval.list) = list1(nod(ODCLFIELD, N, typenod((yyvsp[(1) - (1)].type))));
	}
    break;

  case 362:
#line 2351 "go.y"
    {
		(yyval.node) = nodlit((yyvsp[(1) - (1)].val));
	}
    break;

  case 363:
#line 2355 "go.y"
    {
		(yyval.node) = nodlit((yyvsp[(2) - (2)].val));
		switch((yyval.node)->val.ctype){
		case CTINT:
		case CTRUNE:
			mpnegfix((yyval.node)->val.u.xval);
			break;
		case CTFLT:
			mpnegflt((yyval.node)->val.u.fval);
			break;
		case CTCPLX:
			mpnegflt(&(yyval.node)->val.u.cval->real);
			mpnegflt(&(yyval.node)->val.u.cval->imag);
			break;
		default:
			yyerror("bad negated constant");
		}
	}
    break;

  case 364:
#line 2374 "go.y"
    {
		(yyval.node) = oldname(pkglookup((yyvsp[(1) - (1)].sym)->name, builtinpkg));
		if((yyval.node)->op != OLITERAL)
			yyerror("bad constant %S", (yyval.node)->sym);
	}
    break;

  case 366:
#line 2383 "go.y"
    {
		if((yyvsp[(2) - (5)].node)->val.ctype == CTRUNE && (yyvsp[(4) - (5)].node)->val.ctype == CTINT) {
			(yyval.node) = (yyvsp[(2) - (5)].node);
			mpaddfixfix((yyvsp[(2) - (5)].node)->val.u.xval, (yyvsp[(4) - (5)].node)->val.u.xval, 0);
			break;
		}
		(yyvsp[(4) - (5)].node)->val.u.cval->real = (yyvsp[(4) - (5)].node)->val.u.cval->imag;
		mpmovecflt(&(yyvsp[(4) - (5)].node)->val.u.cval->imag, 0.0);
		(yyval.node) = nodcplxlit((yyvsp[(2) - (5)].node)->val, (yyvsp[(4) - (5)].node)->val);
	}
    break;

  case 369:
#line 2399 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 370:
#line 2403 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 371:
#line 2409 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 372:
#line 2413 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 373:
#line 2419 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 374:
#line 2423 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;


/* Line 1267 of yacc.c.  */
#line 5195 "y.tab.c"
      default: break;
    }
  YY_SYMBOL_PRINT ("-> $$ =", yyr1[yyn], &yyval, &yyloc);

  YYPOPSTACK (yylen);
  yylen = 0;
  YY_STACK_PRINT (yyss, yyssp);

  *++yyvsp = yyval;


  /* Now `shift' the result of the reduction.  Determine what state
     that goes to, based on the state we popped back to and the rule
     number reduced by.  */

  yyn = yyr1[yyn];

  yystate = yypgoto[yyn - YYNTOKENS] + *yyssp;
  if (0 <= yystate && yystate <= YYLAST && yycheck[yystate] == *yyssp)
    yystate = yytable[yystate];
  else
    yystate = yydefgoto[yyn - YYNTOKENS];

  goto yynewstate;


/*------------------------------------.
| yyerrlab -- here on detecting error |
`------------------------------------*/
yyerrlab:
  /* If not already recovering from an error, report this error.  */
  if (!yyerrstatus)
    {
      ++yynerrs;
#if ! YYERROR_VERBOSE
      yyerror (YY_("syntax error"));
#else
      {
	YYSIZE_T yysize = yysyntax_error (0, yystate, yychar);
	if (yymsg_alloc < yysize && yymsg_alloc < YYSTACK_ALLOC_MAXIMUM)
	  {
	    YYSIZE_T yyalloc = 2 * yysize;
	    if (! (yysize <= yyalloc && yyalloc <= YYSTACK_ALLOC_MAXIMUM))
	      yyalloc = YYSTACK_ALLOC_MAXIMUM;
	    if (yymsg != yymsgbuf)
	      YYSTACK_FREE (yymsg);
	    yymsg = (char *) YYSTACK_ALLOC (yyalloc);
	    if (yymsg)
	      yymsg_alloc = yyalloc;
	    else
	      {
		yymsg = yymsgbuf;
		yymsg_alloc = sizeof yymsgbuf;
	      }
	  }

	if (0 < yysize && yysize <= yymsg_alloc)
	  {
	    (void) yysyntax_error (yymsg, yystate, yychar);
	    yyerror (yymsg);
	  }
	else
	  {
	    yyerror (YY_("syntax error"));
	    if (yysize != 0)
	      goto yyexhaustedlab;
	  }
      }
#endif
    }



  if (yyerrstatus == 3)
    {
      /* If just tried and failed to reuse look-ahead token after an
	 error, discard it.  */

      if (yychar <= YYEOF)
	{
	  /* Return failure if at end of input.  */
	  if (yychar == YYEOF)
	    YYABORT;
	}
      else
	{
	  yydestruct ("Error: discarding",
		      yytoken, &yylval);
	  yychar = YYEMPTY;
	}
    }

  /* Else will try to reuse look-ahead token after shifting the error
     token.  */
  goto yyerrlab1;


/*---------------------------------------------------.
| yyerrorlab -- error raised explicitly by YYERROR.  |
`---------------------------------------------------*/
yyerrorlab:

  /* Pacify compilers like GCC when the user code never invokes
     YYERROR and the label yyerrorlab therefore never appears in user
     code.  */
  if (/*CONSTCOND*/ 0)
     goto yyerrorlab;

  /* Do not reclaim the symbols of the rule which action triggered
     this YYERROR.  */
  YYPOPSTACK (yylen);
  yylen = 0;
  YY_STACK_PRINT (yyss, yyssp);
  yystate = *yyssp;
  goto yyerrlab1;


/*-------------------------------------------------------------.
| yyerrlab1 -- common code for both syntax error and YYERROR.  |
`-------------------------------------------------------------*/
yyerrlab1:
  yyerrstatus = 3;	/* Each real token shifted decrements this.  */

  for (;;)
    {
      yyn = yypact[yystate];
      if (yyn != YYPACT_NINF)
	{
	  yyn += YYTERROR;
	  if (0 <= yyn && yyn <= YYLAST && yycheck[yyn] == YYTERROR)
	    {
	      yyn = yytable[yyn];
	      if (0 < yyn)
		break;
	    }
	}

      /* Pop the current state because it cannot handle the error token.  */
      if (yyssp == yyss)
	YYABORT;


      yydestruct ("Error: popping",
		  yystos[yystate], yyvsp);
      YYPOPSTACK (1);
      yystate = *yyssp;
      YY_STACK_PRINT (yyss, yyssp);
    }

  if (yyn == YYFINAL)
    YYACCEPT;

  *++yyvsp = yylval;


  /* Shift the error token.  */
  YY_SYMBOL_PRINT ("Shifting", yystos[yyn], yyvsp, yylsp);

  yystate = yyn;
  goto yynewstate;


/*-------------------------------------.
| yyacceptlab -- YYACCEPT comes here.  |
`-------------------------------------*/
yyacceptlab:
  yyresult = 0;
  goto yyreturn;

/*-----------------------------------.
| yyabortlab -- YYABORT comes here.  |
`-----------------------------------*/
yyabortlab:
  yyresult = 1;
  goto yyreturn;

#ifndef yyoverflow
/*-------------------------------------------------.
| yyexhaustedlab -- memory exhaustion comes here.  |
`-------------------------------------------------*/
yyexhaustedlab:
  yyerror (YY_("memory exhausted"));
  yyresult = 2;
  /* Fall through.  */
#endif

yyreturn:
  if (yychar != YYEOF && yychar != YYEMPTY)
     yydestruct ("Cleanup: discarding lookahead",
		 yytoken, &yylval);
  /* Do not reclaim the symbols of the rule which action triggered
     this YYABORT or YYACCEPT.  */
  YYPOPSTACK (yylen);
  YY_STACK_PRINT (yyss, yyssp);
  while (yyssp != yyss)
    {
      yydestruct ("Cleanup: popping",
		  yystos[*yyssp], yyvsp);
      YYPOPSTACK (1);
    }
#ifndef yyoverflow
  if (yyss != yyssa)
    YYSTACK_FREE (yyss);
#endif
#if YYERROR_VERBOSE
  if (yymsg != yymsgbuf)
    YYSTACK_FREE (yymsg);
#endif
  /* Make sure YYID is used.  */
  return YYID (yyresult);
}


#line 2427 "go.y"


static void
fixlbrace(int lbr)
{
	// If the opening brace was an LBODY,
	// set up for another one now that we're done.
	// See comment in lex.c about loophack.
	if(lbr == LBODY)
		loophack = 1;
}



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
#define YYFINAL  5
/* YYLAST -- Last index in YYTABLE.  */
#define YYLAST   2365

/* YYNTOKENS -- Number of terminals.  */
#define YYNTOKENS  76
/* YYNNTS -- Number of nonterminals.  */
#define YYNNTS  165
/* YYNRULES -- Number of rules.  */
#define YYNRULES  375
/* YYNRULES -- Number of states.  */
#define YYNSTATES  714

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
       0,     0,     3,     8,    21,    22,    26,    27,    31,    32,
      36,    37,    41,    42,    46,    47,    51,    52,    56,    57,
      61,    62,    66,    67,    71,    72,    76,    77,    81,    82,
      86,    87,    91,    94,   100,   104,   108,   111,   113,   117,
     119,   122,   125,   130,   131,   133,   134,   139,   140,   142,
     144,   146,   148,   151,   157,   161,   164,   170,   178,   182,
     185,   191,   195,   197,   200,   205,   209,   214,   218,   220,
     223,   225,   227,   230,   232,   236,   240,   244,   247,   250,
     254,   260,   266,   269,   270,   275,   276,   280,   281,   284,
     285,   290,   295,   300,   303,   309,   311,   313,   316,   317,
     321,   323,   327,   328,   329,   330,   339,   340,   346,   347,
     350,   351,   354,   355,   356,   364,   365,   371,   373,   377,
     381,   385,   389,   393,   397,   401,   405,   409,   413,   417,
     421,   425,   429,   433,   437,   441,   445,   449,   453,   455,
     458,   461,   464,   467,   470,   473,   476,   479,   483,   489,
     496,   498,   500,   504,   510,   516,   521,   528,   537,   539,
     545,   551,   557,   565,   567,   568,   572,   574,   579,   581,
     586,   588,   592,   594,   596,   598,   600,   602,   604,   606,
     607,   609,   611,   613,   615,   620,   625,   627,   629,   631,
     634,   636,   638,   640,   642,   644,   648,   650,   652,   654,
     657,   659,   661,   663,   665,   669,   671,   673,   675,   677,
     679,   681,   683,   685,   687,   691,   696,   701,   704,   708,
     714,   716,   718,   721,   725,   731,   735,   741,   745,   749,
     755,   764,   770,   779,   785,   786,   790,   791,   793,   797,
     799,   804,   807,   808,   812,   814,   818,   820,   824,   826,
     830,   832,   836,   838,   842,   846,   849,   854,   858,   864,
     870,   872,   876,   878,   881,   883,   887,   892,   894,   897,
     900,   902,   904,   908,   909,   912,   913,   915,   917,   919,
     921,   923,   925,   927,   929,   931,   932,   937,   939,   942,
     945,   948,   951,   954,   957,   959,   963,   965,   969,   971,
     975,   977,   981,   983,   987,   989,   991,   995,   999,  1000,
    1003,  1004,  1006,  1007,  1009,  1010,  1012,  1013,  1015,  1016,
    1018,  1019,  1021,  1022,  1024,  1025,  1027,  1028,  1030,  1035,
    1040,  1046,  1053,  1058,  1063,  1065,  1067,  1069,  1071,  1073,
    1075,  1077,  1079,  1081,  1085,  1090,  1096,  1101,  1106,  1109,
    1112,  1117,  1121,  1125,  1131,  1135,  1140,  1144,  1150,  1152,
    1153,  1155,  1159,  1161,  1163,  1166,  1168,  1170,  1176,  1177,
    1180,  1182,  1186,  1188,  1192,  1194
};

/* YYRHS -- A `-1'-separated list of the rules' RHS.  */
static const yytype_int16 yyrhs[] =
{
      77,     0,    -1,    78,    79,   104,   189,    -1,    80,   100,
     102,    90,    82,    92,    94,    98,    96,    84,    86,    88,
      -1,    -1,    25,   164,    62,    -1,    -1,    81,   109,   111,
      -1,    -1,    83,   109,   111,    -1,    -1,    85,   109,   111,
      -1,    -1,    87,   109,   111,    -1,    -1,    89,   109,   111,
      -1,    -1,    91,   109,   111,    -1,    -1,    93,   109,   111,
      -1,    -1,    95,   109,   111,    -1,    -1,    97,   109,   111,
      -1,    -1,    99,   109,   111,    -1,    -1,   101,   109,   111,
      -1,    -1,   103,   109,   111,    -1,    -1,   104,   105,    62,
      -1,    21,   106,    -1,    21,    59,   107,   213,    60,    -1,
      21,    59,    60,    -1,   108,   109,   111,    -1,   108,   111,
      -1,   106,    -1,   107,    62,   106,    -1,     3,    -1,   164,
       3,    -1,    63,     3,    -1,    25,    24,   110,    62,    -1,
      -1,    24,    -1,    -1,   112,   237,    64,    64,    -1,    -1,
     114,    -1,   181,    -1,   204,    -1,     1,    -1,    32,   116,
      -1,    32,    59,   190,   213,    60,    -1,    32,    59,    60,
      -1,   115,   117,    -1,   115,    59,   117,   213,    60,    -1,
     115,    59,   117,    62,   191,   213,    60,    -1,   115,    59,
      60,    -1,    31,   120,    -1,    31,    59,   192,   213,    60,
      -1,    31,    59,    60,    -1,     9,    -1,   208,   169,    -1,
     208,   169,    65,   209,    -1,   208,    65,   209,    -1,   208,
     169,    65,   209,    -1,   208,    65,   209,    -1,   117,    -1,
     208,   169,    -1,   208,    -1,   164,    -1,   119,   169,    -1,
     149,    -1,   149,     4,   149,    -1,   209,    65,   209,    -1,
     209,     5,   209,    -1,   149,    42,    -1,   149,    37,    -1,
       7,   210,    66,    -1,     7,   210,    65,   149,    66,    -1,
       7,   210,     5,   149,    66,    -1,    12,    66,    -1,    -1,
      67,   124,   206,    68,    -1,    -1,   122,   126,   206,    -1,
      -1,   127,   125,    -1,    -1,    35,   129,   206,    68,    -1,
     209,    65,    26,   149,    -1,   209,     5,    26,   149,    -1,
      26,   149,    -1,   217,    62,   217,    62,   217,    -1,   217,
      -1,   130,    -1,   131,   128,    -1,    -1,    16,   134,   132,
      -1,   217,    -1,   217,    62,   217,    -1,    -1,    -1,    -1,
      20,   137,   135,   138,   128,   139,   142,   143,    -1,    -1,
      14,    20,   141,   135,   128,    -1,    -1,   142,   140,    -1,
      -1,    14,   123,    -1,    -1,    -1,    30,   145,   135,   146,
      35,   127,    68,    -1,    -1,    28,   148,    35,   127,    68,
      -1,   150,    -1,   149,    47,   149,    -1,   149,    33,   149,
      -1,   149,    38,   149,    -1,   149,    46,   149,    -1,   149,
      45,   149,    -1,   149,    43,   149,    -1,   149,    39,   149,
      -1,   149,    40,   149,    -1,   149,    49,   149,    -1,   149,
      50,   149,    -1,   149,    51,   149,    -1,   149,    52,   149,
      -1,   149,    53,   149,    -1,   149,    54,   149,    -1,   149,
      55,   149,    -1,   149,    56,   149,    -1,   149,    34,   149,
      -1,   149,    44,   149,    -1,   149,    48,   149,    -1,   149,
      36,   149,    -1,   157,    -1,    53,   150,    -1,    56,   150,
      -1,    49,   150,    -1,    50,   150,    -1,    69,   150,    -1,
      70,   150,    -1,    52,   150,    -1,    36,   150,    -1,   157,
      59,    60,    -1,   157,    59,   210,   214,    60,    -1,   157,
      59,   210,    11,   214,    60,    -1,     3,    -1,   166,    -1,
     157,    63,   164,    -1,   157,    63,    59,   158,    60,    -1,
     157,    63,    59,    31,    60,    -1,   157,    71,   149,    72,
      -1,   157,    71,   215,    66,   215,    72,    -1,   157,    71,
     215,    66,   215,    66,   215,    72,    -1,   151,    -1,   172,
      59,   149,   214,    60,    -1,   173,   160,   153,   212,    68,
      -1,   152,    67,   153,   212,    68,    -1,    59,   158,    60,
      67,   153,   212,    68,    -1,   188,    -1,    -1,   149,    66,
     156,    -1,   149,    -1,    67,   153,   212,    68,    -1,   149,
      -1,    67,   153,   212,    68,    -1,   152,    -1,    59,   158,
      60,    -1,   149,    -1,   170,    -1,   169,    -1,    35,    -1,
      67,    -1,   164,    -1,   164,    -1,    -1,   161,    -1,    24,
      -1,   165,    -1,    73,    -1,    74,     3,    63,    24,    -1,
      74,     3,    63,    73,    -1,   164,    -1,   161,    -1,    11,
      -1,    11,   169,    -1,   178,    -1,   184,    -1,   176,    -1,
     177,    -1,   175,    -1,    59,   169,    60,    -1,   178,    -1,
     184,    -1,   176,    -1,    53,   170,    -1,   184,    -1,   176,
      -1,   177,    -1,   175,    -1,    59,   169,    60,    -1,   184,
      -1,   176,    -1,   176,    -1,   178,    -1,   184,    -1,   176,
      -1,   177,    -1,   175,    -1,   166,    -1,   166,    63,   164,
      -1,    71,   215,    72,   169,    -1,    71,    11,    72,   169,
      -1,     8,   171,    -1,     8,    36,   169,    -1,    23,    71,
     169,    72,   169,    -1,   179,    -1,   180,    -1,    53,   169,
      -1,    36,     8,   169,    -1,    29,   160,   193,   213,    68,
      -1,    29,   160,    68,    -1,    22,   160,   194,   213,    68,
      -1,    22,   160,    68,    -1,    17,   182,   185,    -1,   164,
      59,   202,    60,   186,    -1,    59,   202,    60,   164,    59,
     202,    60,   186,    -1,   223,    59,   218,    60,   233,    -1,
      59,   238,    60,   164,    59,   218,    60,   233,    -1,    17,
      59,   202,    60,   186,    -1,    -1,    67,   206,    68,    -1,
      -1,   174,    -1,    59,   202,    60,    -1,   184,    -1,   187,
     160,   206,    68,    -1,   187,     1,    -1,    -1,   189,   113,
      62,    -1,   116,    -1,   190,    62,   116,    -1,   118,    -1,
     191,    62,   118,    -1,   120,    -1,   192,    62,   120,    -1,
     195,    -1,   193,    62,   195,    -1,   198,    -1,   194,    62,
     198,    -1,   207,   169,   221,    -1,   197,   221,    -1,    59,
     197,    60,   221,    -1,    53,   197,   221,    -1,    59,    53,
     197,    60,   221,    -1,    53,    59,   197,    60,   221,    -1,
      24,    -1,    24,    63,   164,    -1,   196,    -1,   161,   199,
      -1,   196,    -1,    59,   196,    60,    -1,    59,   202,    60,
     186,    -1,   159,    -1,   164,   159,    -1,   164,   168,    -1,
     168,    -1,   200,    -1,   201,    75,   200,    -1,    -1,   201,
     214,    -1,    -1,   123,    -1,   114,    -1,   204,    -1,     1,
      -1,   121,    -1,   133,    -1,   144,    -1,   147,    -1,   136,
      -1,    -1,   167,    66,   205,   203,    -1,    15,    -1,     6,
     163,    -1,    10,   163,    -1,    18,   151,    -1,    13,   151,
      -1,    19,   161,    -1,    27,   216,    -1,   203,    -1,   206,
      62,   203,    -1,   161,    -1,   207,    75,   161,    -1,   162,
      -1,   208,    75,   162,    -1,   149,    -1,   209,    75,   149,
      -1,   158,    -1,   210,    75,   158,    -1,   154,    -1,   155,
      -1,   211,    75,   154,    -1,   211,    75,   155,    -1,    -1,
     211,   214,    -1,    -1,    62,    -1,    -1,    75,    -1,    -1,
     149,    -1,    -1,   209,    -1,    -1,   121,    -1,    -1,   238,
      -1,    -1,   239,    -1,    -1,   240,    -1,    -1,     3,    -1,
      21,    24,     3,    62,    -1,    32,   223,   225,    62,    -1,
       9,   223,    65,   236,    62,    -1,     9,   223,   225,    65,
     236,    62,    -1,    31,   224,   225,    62,    -1,    17,   183,
     185,    62,    -1,   165,    -1,   223,    -1,   227,    -1,   228,
      -1,   229,    -1,   227,    -1,   229,    -1,   165,    -1,    24,
      -1,    71,    72,   225,    -1,    71,     3,    72,   225,    -1,
      23,    71,   225,    72,   225,    -1,    29,    67,   219,    68,
      -1,    22,    67,   220,    68,    -1,    53,   225,    -1,     8,
     226,    -1,     8,    59,   228,    60,    -1,     8,    36,   225,
      -1,    36,     8,   225,    -1,    17,    59,   218,    60,   233,
      -1,   164,   225,   221,    -1,   164,    11,   225,   221,    -1,
     164,   225,   221,    -1,   164,    59,   218,    60,   233,    -1,
     225,    -1,    -1,   234,    -1,    59,   218,    60,    -1,   225,
      -1,     3,    -1,    50,     3,    -1,   164,    -1,   235,    -1,
      59,   235,    49,   235,    60,    -1,    -1,   237,   222,    -1,
     230,    -1,   238,    75,   230,    -1,   231,    -1,   239,    62,
     231,    -1,   232,    -1,   240,    62,   232,    -1
};

/* YYRLINE[YYN] -- source line where rule number YYN was defined.  */
static const yytype_uint16 yyrline[] =
{
       0,   124,   124,   133,   149,   155,   166,   166,   183,   183,
     201,   201,   219,   219,   237,   237,   255,   255,   273,   273,
     290,   290,   308,   308,   326,   326,   344,   344,   367,   367,
     383,   384,   387,   388,   389,   392,   429,   440,   441,   444,
     451,   458,   467,   481,   482,   489,   489,   502,   506,   507,
     511,   516,   522,   526,   530,   534,   540,   546,   552,   557,
     561,   565,   571,   577,   581,   585,   591,   595,   601,   602,
     606,   612,   621,   627,   645,   650,   662,   678,   684,   692,
     712,   730,   739,   758,   757,   772,   771,   803,   806,   813,
     812,   823,   829,   836,   843,   854,   860,   863,   871,   870,
     881,   887,   899,   903,   908,   898,   929,   928,   941,   944,
     950,   953,   965,   969,   964,   987,   986,  1002,  1003,  1007,
    1011,  1015,  1019,  1023,  1027,  1031,  1035,  1039,  1043,  1047,
    1051,  1055,  1059,  1063,  1067,  1071,  1075,  1080,  1086,  1087,
    1091,  1102,  1106,  1110,  1114,  1119,  1123,  1133,  1137,  1142,
    1150,  1154,  1155,  1166,  1170,  1174,  1178,  1182,  1190,  1191,
    1197,  1204,  1210,  1217,  1220,  1227,  1233,  1250,  1257,  1258,
    1265,  1266,  1285,  1286,  1289,  1292,  1296,  1307,  1316,  1322,
    1325,  1328,  1335,  1336,  1342,  1355,  1370,  1378,  1390,  1395,
    1401,  1402,  1403,  1404,  1405,  1406,  1412,  1413,  1414,  1415,
    1421,  1422,  1423,  1424,  1425,  1431,  1432,  1435,  1438,  1439,
    1440,  1441,  1442,  1445,  1446,  1459,  1463,  1468,  1473,  1478,
    1482,  1483,  1486,  1492,  1499,  1505,  1512,  1518,  1529,  1545,
    1574,  1612,  1637,  1655,  1664,  1667,  1675,  1679,  1683,  1690,
    1696,  1701,  1713,  1716,  1728,  1729,  1735,  1736,  1742,  1746,
    1752,  1753,  1759,  1763,  1769,  1792,  1797,  1803,  1809,  1816,
    1825,  1834,  1849,  1855,  1860,  1864,  1871,  1884,  1885,  1891,
    1897,  1900,  1904,  1910,  1913,  1922,  1925,  1926,  1930,  1931,
    1937,  1938,  1939,  1940,  1941,  1943,  1942,  1957,  1963,  1967,
    1971,  1975,  1979,  1984,  2003,  2009,  2017,  2021,  2027,  2031,
    2037,  2041,  2047,  2051,  2060,  2064,  2068,  2072,  2078,  2081,
    2089,  2090,  2092,  2093,  2096,  2099,  2102,  2105,  2108,  2111,
    2114,  2117,  2120,  2123,  2126,  2129,  2132,  2135,  2141,  2145,
    2149,  2153,  2157,  2161,  2181,  2188,  2199,  2200,  2201,  2204,
    2205,  2208,  2212,  2222,  2226,  2230,  2234,  2238,  2242,  2246,
    2252,  2258,  2266,  2274,  2280,  2287,  2303,  2325,  2329,  2335,
    2338,  2341,  2345,  2355,  2359,  2378,  2386,  2387,  2399,  2400,
    2403,  2407,  2413,  2417,  2423,  2427
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
  "loadsys", "package", "loadcore", "@1", "loadchannels", "@2", "loadseq",
  "@3", "loaddefers", "@4", "loadruntime", "@5", "loadsched", "@6",
  "loadhash", "@7", "loadprintf", "@8", "loadifacestuff", "@9",
  "loadstrings", "@10", "loadmaps", "@11", "loadwb", "@12", "imports",
  "import", "import_stmt", "import_stmt_list", "import_here",
  "import_package", "import_safety", "import_there", "@13", "xdcl",
  "common_dcl", "lconst", "vardcl", "constdcl", "constdcl1", "typedclname",
  "typedcl", "simple_stmt", "case", "compound_stmt", "@14", "caseblock",
  "@15", "caseblock_list", "loop_body", "@16", "range_stmt", "for_header",
  "for_body", "for_stmt", "@17", "if_header", "if_stmt", "@18", "@19",
  "@20", "elseif", "@21", "elseif_list", "else", "switch_stmt", "@22",
  "@23", "select_stmt", "@24", "expr", "uexpr", "pseudocall",
  "pexpr_no_paren", "start_complit", "keyval", "bare_complitexpr",
  "complitexpr", "pexpr", "expr_or_type", "name_or_type", "lbrace",
  "new_name", "dcl_name", "onew_name", "sym", "hidden_importsym", "name",
  "labelname", "dotdotdot", "ntype", "non_expr_type", "non_recvchantype",
  "convtype", "comptype", "fnret_type", "dotname", "othertype", "ptrtype",
  "recvchantype", "structtype", "interfacetype", "xfndcl", "fndcl",
  "hidden_fndcl", "fntype", "fnbody", "fnres", "fnlitdcl", "fnliteral",
  "xdcl_list", "vardcl_list", "constdcl_list", "typedcl_list",
  "structdcl_list", "interfacedcl_list", "structdcl", "packname", "embed",
  "interfacedcl", "indcl", "arg_type", "arg_type_list",
  "oarg_type_list_ocomma", "stmt", "non_dcl_stmt", "@25", "stmt_list",
  "new_name_list", "dcl_name_list", "expr_list", "expr_or_type_list",
  "keyval_list", "braced_keyval_list", "osemi", "ocomma", "oexpr",
  "oexpr_list", "osimple_stmt", "ohidden_funarg_list",
  "ohidden_structdcl_list", "ohidden_interfacedcl_list", "oliteral",
  "hidden_import", "hidden_pkg_importsym", "hidden_pkgtype", "hidden_type",
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
       0,    76,    77,    78,    79,    79,    81,    80,    83,    82,
      85,    84,    87,    86,    89,    88,    91,    90,    93,    92,
      95,    94,    97,    96,    99,    98,   101,   100,   103,   102,
     104,   104,   105,   105,   105,   106,   106,   107,   107,   108,
     108,   108,   109,   110,   110,   112,   111,   113,   113,   113,
     113,   113,   114,   114,   114,   114,   114,   114,   114,   114,
     114,   114,   115,   116,   116,   116,   117,   117,   118,   118,
     118,   119,   120,   121,   121,   121,   121,   121,   121,   122,
     122,   122,   122,   124,   123,   126,   125,   127,   127,   129,
     128,   130,   130,   130,   131,   131,   131,   132,   134,   133,
     135,   135,   137,   138,   139,   136,   141,   140,   142,   142,
     143,   143,   145,   146,   144,   148,   147,   149,   149,   149,
     149,   149,   149,   149,   149,   149,   149,   149,   149,   149,
     149,   149,   149,   149,   149,   149,   149,   149,   150,   150,
     150,   150,   150,   150,   150,   150,   150,   151,   151,   151,
     152,   152,   152,   152,   152,   152,   152,   152,   152,   152,
     152,   152,   152,   152,   153,   154,   155,   155,   156,   156,
     157,   157,   158,   158,   159,   160,   160,   161,   162,   163,
     163,   164,   164,   164,   165,   165,   166,   167,   168,   168,
     169,   169,   169,   169,   169,   169,   170,   170,   170,   170,
     171,   171,   171,   171,   171,   172,   172,   173,   174,   174,
     174,   174,   174,   175,   175,   176,   176,   176,   176,   176,
     176,   176,   177,   178,   179,   179,   180,   180,   181,   182,
     182,   183,   183,   184,   185,   185,   186,   186,   186,   187,
     188,   188,   189,   189,   190,   190,   191,   191,   192,   192,
     193,   193,   194,   194,   195,   195,   195,   195,   195,   195,
     196,   196,   197,   198,   198,   198,   199,   200,   200,   200,
     200,   201,   201,   202,   202,   203,   203,   203,   203,   203,
     204,   204,   204,   204,   204,   205,   204,   204,   204,   204,
     204,   204,   204,   204,   206,   206,   207,   207,   208,   208,
     209,   209,   210,   210,   211,   211,   211,   211,   212,   212,
     213,   213,   214,   214,   215,   215,   216,   216,   217,   217,
     218,   218,   219,   219,   220,   220,   221,   221,   222,   222,
     222,   222,   222,   222,   223,   224,   225,   225,   225,   226,
     226,   227,   227,   227,   227,   227,   227,   227,   227,   227,
     227,   227,   228,   229,   230,   230,   231,   232,   232,   233,
     233,   234,   234,   235,   235,   235,   236,   236,   237,   237,
     238,   238,   239,   239,   240,   240
};

/* YYR2[YYN] -- Number of symbols composing right hand side of rule YYN.  */
static const yytype_uint8 yyr2[] =
{
       0,     2,     4,    12,     0,     3,     0,     3,     0,     3,
       0,     3,     0,     3,     0,     3,     0,     3,     0,     3,
       0,     3,     0,     3,     0,     3,     0,     3,     0,     3,
       0,     3,     2,     5,     3,     3,     2,     1,     3,     1,
       2,     2,     4,     0,     1,     0,     4,     0,     1,     1,
       1,     1,     2,     5,     3,     2,     5,     7,     3,     2,
       5,     3,     1,     2,     4,     3,     4,     3,     1,     2,
       1,     1,     2,     1,     3,     3,     3,     2,     2,     3,
       5,     5,     2,     0,     4,     0,     3,     0,     2,     0,
       4,     4,     4,     2,     5,     1,     1,     2,     0,     3,
       1,     3,     0,     0,     0,     8,     0,     5,     0,     2,
       0,     2,     0,     0,     7,     0,     5,     1,     3,     3,
       3,     3,     3,     3,     3,     3,     3,     3,     3,     3,
       3,     3,     3,     3,     3,     3,     3,     3,     1,     2,
       2,     2,     2,     2,     2,     2,     2,     3,     5,     6,
       1,     1,     3,     5,     5,     4,     6,     8,     1,     5,
       5,     5,     7,     1,     0,     3,     1,     4,     1,     4,
       1,     3,     1,     1,     1,     1,     1,     1,     1,     0,
       1,     1,     1,     1,     4,     4,     1,     1,     1,     2,
       1,     1,     1,     1,     1,     3,     1,     1,     1,     2,
       1,     1,     1,     1,     3,     1,     1,     1,     1,     1,
       1,     1,     1,     1,     3,     4,     4,     2,     3,     5,
       1,     1,     2,     3,     5,     3,     5,     3,     3,     5,
       8,     5,     8,     5,     0,     3,     0,     1,     3,     1,
       4,     2,     0,     3,     1,     3,     1,     3,     1,     3,
       1,     3,     1,     3,     3,     2,     4,     3,     5,     5,
       1,     3,     1,     2,     1,     3,     4,     1,     2,     2,
       1,     1,     3,     0,     2,     0,     1,     1,     1,     1,
       1,     1,     1,     1,     1,     0,     4,     1,     2,     2,
       2,     2,     2,     2,     1,     3,     1,     3,     1,     3,
       1,     3,     1,     3,     1,     1,     3,     3,     0,     2,
       0,     1,     0,     1,     0,     1,     0,     1,     0,     1,
       0,     1,     0,     1,     0,     1,     0,     1,     4,     4,
       5,     6,     4,     4,     1,     1,     1,     1,     1,     1,
       1,     1,     1,     3,     4,     5,     4,     4,     2,     2,
       4,     3,     3,     5,     3,     4,     3,     5,     1,     0,
       1,     3,     1,     1,     2,     1,     1,     5,     0,     2,
       1,     3,     1,     3,     1,     3
};

/* YYDEFACT[STATE-NAME] -- Default rule to reduce with in state
   STATE-NUM when YYTABLE doesn't specify something else to do.  Zero
   means the default is an error.  */
static const yytype_uint16 yydefact[] =
{
       6,     0,     4,    26,     0,     1,     0,    30,    28,     0,
       0,    45,   181,   183,     0,     0,   182,   242,    16,     0,
      45,    43,     7,   368,     0,     5,     0,     0,     0,     8,
       0,    45,    27,    44,     0,     0,     0,    39,     0,     0,
      32,    45,     0,    31,    51,   150,   179,     0,    62,   179,
       0,   287,    98,     0,     0,     0,   102,     0,     0,   316,
     115,     0,   112,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,   314,     0,    48,     0,   280,   281,
     284,   282,   283,    73,   117,   158,   170,   138,   187,   186,
     151,     0,     0,     0,   207,   220,   221,    49,   239,     0,
     163,    50,     0,    18,     0,    45,    29,    42,     0,     0,
       0,     0,     0,     0,   369,   184,   185,    34,    37,   310,
      41,    45,    36,    40,   180,   288,   177,     0,     0,     0,
       0,   186,   213,   217,   203,   201,   202,   200,   289,   158,
       0,   318,   273,     0,   234,   158,   292,   318,   175,   176,
       0,     0,   300,   317,   293,     0,     0,   318,     0,     0,
      59,    71,     0,    52,   298,   178,     0,   146,   141,   142,
     145,   139,   140,     0,     0,   172,     0,   173,   198,   196,
     197,   143,   144,     0,   315,     0,   243,     0,    55,     0,
       0,     0,     0,     0,    78,     0,     0,     0,    77,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,   164,     0,     0,   314,   285,     0,   164,
     241,     0,     0,     0,     0,    20,     0,    45,    17,   334,
       0,     0,   234,     0,     0,   335,     0,     0,    46,   311,
       0,    35,   273,     0,     0,   218,   194,   192,   193,   190,
     191,   222,     0,     0,     0,   319,    96,     0,    99,     0,
      95,   188,   267,   186,   270,   174,   271,   312,     0,   273,
       0,   228,   103,   100,   181,     0,   227,     0,   310,   264,
     252,     0,    87,     0,     0,   225,   296,   310,   250,   262,
     326,     0,   113,    61,   248,   310,    72,    54,   244,   310,
       0,     0,    63,     0,   199,   171,     0,     0,    58,   310,
       0,     0,    74,   119,   134,   137,   120,   124,   125,   123,
     135,   122,   121,   118,   136,   126,   127,   128,   129,   130,
     131,   132,   133,   308,   147,   302,   312,     0,   152,   315,
       0,     0,   312,   308,   279,    83,   277,   276,   294,   278,
       0,    76,    75,   301,    24,     0,    45,     9,     0,     0,
       0,     0,   342,     0,     0,     0,     0,     0,   341,     0,
     336,   337,   338,     0,   370,     0,     0,   320,     0,     0,
       0,    38,    33,     0,     0,     0,   204,   214,    93,    89,
      97,     0,     0,   318,   189,   268,   269,   313,   274,   236,
       0,     0,     0,   318,     0,   260,     0,   273,   263,   311,
       0,     0,     0,     0,   326,     0,     0,   311,     0,   327,
     255,     0,   326,     0,   311,     0,   311,     0,    65,   299,
       0,     0,     0,   223,   194,   192,   193,   191,   164,   216,
     215,   311,     0,    67,     0,   164,   166,   304,   305,   312,
       0,   312,   313,     0,     0,     0,   155,   314,   286,   313,
       0,     0,     0,     0,   240,    22,     0,    45,    19,     0,
       0,   349,   339,   340,   320,   324,     0,   322,     0,   348,
     363,     0,     0,   365,   366,     0,     0,     0,     0,     0,
     326,     0,     0,   333,     0,   321,   328,   332,   329,   236,
     195,     0,     0,     0,     0,   272,   273,   186,   237,   212,
     210,   211,   208,   209,   233,   236,   235,   104,   101,   261,
     265,     0,   253,   226,   219,     0,     0,   116,    85,    88,
       0,   257,     0,   326,   251,   224,   297,   254,    87,   249,
      60,   245,    53,    64,     0,   308,    68,   246,   310,    70,
      56,    66,   308,     0,   313,   309,   161,     0,   303,   148,
     154,   153,     0,   159,   160,     0,   295,    10,     0,    45,
      21,   351,     0,     0,   342,     0,   341,     0,   358,   374,
     325,     0,     0,     0,   372,   323,   352,   364,     0,   330,
       0,   343,     0,   326,   354,     0,   371,   359,     0,    92,
      91,   318,     0,   273,   229,   108,   236,     0,    82,     0,
     326,   326,   256,     0,   195,     0,   311,     0,    69,     0,
     164,   168,   165,   306,   307,   149,   314,   156,    84,    12,
       0,    45,    25,   350,   359,   320,   347,     0,     0,   326,
     346,     0,     0,   344,   331,   355,   320,   320,   362,   231,
     360,    90,    94,   238,     0,   110,   266,     0,     0,    79,
       0,    86,   259,   258,   114,   162,   247,    57,   167,   308,
       0,    14,     0,    45,    23,   353,     0,   375,   345,   356,
     373,     0,     0,     0,   236,     0,   109,   105,     0,     0,
       0,   157,     3,     0,    45,    11,   359,   367,   359,   361,
     230,   106,   111,    81,    80,   169,    45,    13,   357,   232,
     318,    15,     0,   107
};

/* YYDEFGOTO[NTERM-NUM].  */
static const yytype_int16 yydefgoto[] =
{
      -1,     1,     2,     7,     3,     4,   103,   104,   629,   630,
     671,   672,   692,   693,    29,    30,   225,   226,   354,   355,
     567,   568,   465,   466,     8,     9,    18,    19,    17,    27,
      40,   119,    41,    11,    34,    22,    23,    75,   346,    77,
     163,   546,   547,   159,   160,    78,   528,   347,   462,   529,
     609,   412,   390,   501,   256,   257,   258,    79,   141,   272,
      80,   147,   402,   605,   686,   710,   655,   687,    81,   157,
     423,    82,   155,    83,    84,    85,    86,   333,   447,   448,
     622,    87,   335,   262,   150,    88,   164,   125,   131,    16,
      90,    91,   264,   265,   177,   133,    92,    93,   508,   246,
      94,   248,   249,    95,    96,    97,   144,   232,    98,   271,
     514,    99,   100,    28,   299,   548,   295,   287,   278,   288,
     289,   290,   280,   408,   266,   267,   268,   348,   349,   341,
     350,   291,   166,   102,   336,   449,   450,   240,   398,   185,
     154,   273,   494,   583,   577,   420,   114,   230,   236,   648,
     471,   370,   371,   372,   374,   584,   579,   649,   650,   484,
     485,    35,   495,   585,   580
};

/* YYPACT[STATE-NUM] -- Index in YYTABLE of the portion describing
   STATE-NUM.  */
#define YYPACT_NINF -578
static const yytype_int16 yypact[] =
{
    -578,    56,    58,  -578,    59,  -578,   135,  -578,  -578,    59,
      46,  -578,  -578,  -578,    79,    38,  -578,    97,  -578,    59,
    -578,   127,  -578,  -578,   104,  -578,    73,   108,  1292,  -578,
      59,  -578,  -578,  -578,   113,   286,    37,  -578,    75,   175,
    -578,    59,   189,  -578,  -578,  -578,   135,  1888,  -578,   135,
      57,  -578,  -578,   213,    57,   135,  -578,   110,   134,  1759,
    -578,   110,  -578,   355,   371,  1759,  1759,  1759,  1759,  1759,
    1759,  1802,  1759,  1759,  1076,   144,  -578,   379,  -578,  -578,
    -578,  -578,  -578,   969,  -578,  -578,   168,   368,  -578,   174,
    -578,   188,   202,   110,   203,  -578,  -578,  -578,   218,    70,
    -578,  -578,   107,  -578,    59,  -578,  -578,  -578,   191,   -12,
     256,   191,   191,   220,  -578,  -578,  -578,  -578,  -578,   223,
    -578,  -578,  -578,  -578,  -578,  -578,  -578,   235,  1913,  1913,
    1913,  -578,   237,  -578,  -578,  -578,  -578,  -578,  -578,   164,
     368,  1145,   917,   243,   239,   248,  -578,  1759,  -578,  -578,
     274,  1913,  2238,   229,  -578,   276,   266,  1759,   362,  1913,
    -578,  -578,   426,  -578,  -578,  -578,   494,  -578,  -578,  -578,
    -578,  -578,  -578,  1845,  1802,  2238,   249,  -578,   270,  -578,
      67,  -578,  -578,   250,  2238,   251,  -578,   430,  -578,   802,
    1759,  1759,  1759,  1759,  -578,  1759,  1759,  1759,  -578,  1759,
    1759,  1759,  1759,  1759,  1759,  1759,  1759,  1759,  1759,  1759,
    1759,  1759,  1759,  -578,  1188,   435,  1759,  -578,  1759,  -578,
    -578,  1441,  1759,  1759,  1759,  -578,    59,  -578,  -578,  -578,
     291,   135,   239,   262,   329,  -578,  2097,  2097,  -578,   100,
     268,  -578,   917,   330,  1913,  -578,  -578,  -578,  -578,  -578,
    -578,  -578,   301,   135,  1759,  -578,  -578,   336,  -578,   120,
     310,  1913,  -578,   917,  -578,  -578,  -578,   305,   323,   917,
    1441,  -578,  -578,   331,   136,   370,  -578,   346,   354,  -578,
    -578,   351,  -578,    45,   128,  -578,  -578,   375,  -578,  -578,
     429,   740,  -578,  -578,  -578,   378,  -578,  -578,  -578,   381,
    1759,   135,   383,  1921,  -578,   366,  1913,  1913,  -578,   384,
    1759,   386,  2238,  2309,  -578,  2262,   672,   672,   672,   672,
    -578,   672,   672,  2286,  -578,   539,   539,   539,   539,  -578,
    -578,  -578,  -578,  1496,  -578,  -578,    34,  1551,  -578,  2136,
     389,  1367,  1231,  1496,  -578,  -578,  -578,  -578,  -578,  -578,
       4,   229,   229,  2238,  -578,    59,  -578,  -578,  2005,   397,
     391,   390,  -578,   400,   463,  2097,   177,    48,  -578,   408,
    -578,  -578,  -578,  2013,  -578,    -2,   414,   135,   416,   421,
     422,  -578,  -578,   438,  1913,   447,  -578,  -578,  2238,  -578,
    -578,  1606,  1661,  1759,  -578,  -578,  -578,   917,  -578,  1946,
     450,    65,   336,  1759,   135,   432,   452,   917,  -578,   446,
     453,  1913,   101,   370,   429,   370,   464,   328,   458,  -578,
    -578,   135,   429,   497,   135,   473,   135,   475,   229,  -578,
    1759,  1980,  1913,  -578,   296,   300,   325,   339,  -578,  -578,
    -578,   135,   476,   229,  1759,  -578,  2166,  -578,  -578,   462,
     470,   465,  1802,   479,   482,   485,  -578,  1759,  -578,  -578,
     488,   483,  1441,  1367,  -578,  -578,    59,  -578,  -578,  2097,
     514,  -578,  -578,  -578,   135,  2038,  2097,   135,  2097,  -578,
    -578,   552,    91,  -578,  -578,   495,   484,  2097,   177,  2097,
     429,   135,   135,  -578,   498,   486,  -578,  -578,  -578,  1946,
    -578,  1441,  1759,  1759,   501,  -578,   917,   505,  -578,  -578,
    -578,  -578,  -578,  -578,  -578,  1946,  -578,  -578,  -578,  -578,
    -578,   500,  -578,  -578,  -578,  1802,   504,  -578,  -578,  -578,
     506,  -578,   512,   429,  -578,  -578,  -578,  -578,  -578,  -578,
    -578,  -578,  -578,   229,   515,  1496,  -578,  -578,   518,   802,
    -578,   229,  1496,  1704,  1496,  -578,  -578,   522,  -578,  -578,
    -578,  -578,   153,  -578,  -578,   166,  -578,  -578,    59,  -578,
    -578,  -578,   524,   525,   540,   542,   543,   536,  -578,  -578,
     544,   533,  2097,   545,  -578,   548,  -578,  -578,   563,  -578,
    2097,  -578,   555,   429,  -578,   559,  -578,  2072,   190,  2238,
    2238,  1759,   554,   917,  -578,  -578,  1946,   118,  -578,  1367,
     429,   429,  -578,   206,   341,   551,   135,   561,   386,   557,
    -578,  2238,  -578,  -578,  -578,  -578,  1759,  -578,  -578,  -578,
      59,  -578,  -578,  -578,  2072,   135,  -578,  2038,  2097,   429,
    -578,   135,    91,  -578,  -578,  -578,   135,   135,  -578,  -578,
    -578,  -578,  -578,  -578,   566,   609,  -578,  1759,  1759,  -578,
    1802,   565,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  1496,
     556,  -578,    59,  -578,  -578,  -578,   569,  -578,  -578,  -578,
    -578,   570,   572,   573,  1946,    39,  -578,  -578,  2190,  2214,
     568,  -578,  -578,    59,  -578,  -578,  2072,  -578,  2072,  -578,
    -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,
    1759,  -578,   336,  -578
};

/* YYPGOTO[NTERM-NUM].  */
static const yytype_int16 yypgoto[] =
{
    -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,
    -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,
    -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,
     -16,  -578,  -578,    -6,  -578,   -20,  -578,  -578,   606,  -578,
    -144,   -27,    24,  -578,  -143,  -114,  -578,   -46,  -578,  -578,
    -578,   106,  -400,  -578,  -578,  -578,  -578,  -578,  -578,  -156,
    -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,
    -578,  -578,  -578,   700,    23,    90,  -578,  -200,    89,    94,
    -578,   193,   -62,   382,   198,   -39,   349,   602,     0,   593,
     449,  -578,   392,   677,   478,  -578,  -578,  -578,  -578,   -35,
      82,   -30,    16,  -578,  -578,  -578,  -578,  -578,   283,   424,
    -484,  -578,  -578,  -578,  -578,  -578,  -578,  -578,  -578,   241,
    -121,  -247,   252,  -578,   263,  -578,  -228,  -302,   631,  -578,
    -218,  -578,   -73,   -34,   137,  -578,  -323,  -255,  -294,  -208,
    -578,  -136,  -444,  -578,  -578,  -347,  -578,   105,  -578,   292,
    -578,   306,   199,   312,   171,    27,    35,  -577,  -578,  -448,
     183,  -578,   442,  -578,  -578
};

/* YYTABLE[YYPACT[STATE-NUM]].  What to do in state STATE-NUM.  If
   positive, shift that token.  If negative, reduce the rule which
   number is the opposite.  If zero, do what YYDEFACT says.
   If YYTABLE_NINF, syntax error.  */
#define YYTABLE_NINF -301
static const yytype_int16 yytable[] =
{
      32,   292,   517,    20,   189,   260,    15,   124,   340,   176,
     124,   106,   134,    31,   383,   294,   146,   136,   298,   343,
     461,   122,   118,   410,   105,   153,    42,   255,    89,   279,
     573,   604,   418,   255,   588,   121,   414,   416,    42,   458,
     425,   400,   453,   255,   427,   451,   126,   231,   460,   126,
     188,   486,   401,   143,   442,   126,     5,   675,   491,   701,
      45,   115,    14,   161,   165,    47,   463,   531,  -239,   405,
      21,   220,   464,   492,   127,   537,    37,   165,    37,    57,
      58,    12,    24,     6,    10,   228,    61,   179,   167,   168,
     169,   170,   171,   172,   480,   181,   182,    12,   227,    12,
      25,   241,  -239,    37,   413,   148,   345,   259,   525,   452,
     116,   277,   222,   526,   189,    12,    71,   286,    26,   708,
     487,   709,   656,   657,    12,   391,  -205,   463,    74,   135,
      13,    14,    38,   516,  -239,   117,    39,   149,    39,  -260,
     139,   481,   263,   594,   145,   148,    13,    14,    13,    14,
     126,    33,   405,   178,   406,   555,   126,   557,   161,    12,
     309,   566,   165,    39,    13,    14,   530,    36,   532,   527,
      43,  -291,   223,    13,    14,   107,  -291,   149,   120,   521,
     480,   415,   224,   658,   659,   392,   612,   165,   351,   352,
     179,   676,   123,   660,   681,   224,   167,   171,  -260,   404,
     700,    12,   682,   683,  -260,   151,   186,   357,    13,    14,
     247,   247,   247,   525,   233,   338,   235,   237,   526,   626,
     356,    89,   615,   381,   247,   627,  -291,   481,   463,   619,
     179,   373,  -291,   247,   628,   213,   482,    12,   545,    42,
    -177,   247,   263,   140,   565,   552,   645,   140,   247,   562,
      13,    14,   463,   387,   217,  -290,   178,   504,   651,   156,
    -290,   218,  -206,   662,   663,    14,   428,   518,   434,   263,
      89,   247,   142,   436,   664,   455,   443,  -205,   602,   255,
     234,   539,   541,   598,   238,   239,    13,    14,   279,   255,
     274,   219,   679,   617,   242,   108,   178,   221,   274,   358,
     253,   165,   269,   109,   224,  -207,   270,   110,   359,   305,
    -290,   282,   713,   360,   361,   362,  -290,   111,   112,   283,
     363,   377,   306,   307,   247,   284,   247,   364,   382,  -206,
     137,  -203,   378,   275,   285,  -201,   468,  -207,   384,    13,
      14,    89,   276,   247,   365,   247,   690,    13,    14,   467,
     113,   247,   274,   179,   180,  -203,   366,   351,   352,  -201,
    -202,   386,   367,  -203,   509,    14,   483,  -201,   549,   511,
     277,   389,   393,   247,  -200,   654,  -204,   373,   286,    12,
     397,   283,   536,   399,  -202,   435,    12,   284,   247,   247,
     558,   661,  -202,   403,   405,    12,   543,   263,  -200,   507,
    -204,    13,    14,    12,   519,   407,  -200,   263,  -204,   126,
     551,   250,   250,   250,   158,   512,   409,   126,   670,   178,
     669,   126,   293,   411,   161,   250,   165,   214,    13,    14,
     162,   215,   419,   438,   250,    13,    14,   417,   187,   216,
     424,   165,   250,   426,    13,    14,   441,   570,   430,   250,
      12,   444,    13,    14,    12,   457,   474,   180,   475,    12,
     569,   476,    89,    89,   509,   652,   247,   477,   179,   511,
     274,   478,   250,   488,   373,   575,   493,   582,   496,   247,
     509,   510,   483,   497,   498,   511,   297,   255,   483,   247,
     308,   595,   373,   247,   337,   404,   132,   180,   499,    13,
      14,    89,    47,    13,    14,   275,   263,   500,    13,    14,
     515,   127,   520,   247,   247,   512,    57,    58,    12,    13,
      14,   523,   369,    61,   533,   250,   535,   250,   379,   380,
     243,   512,   538,   540,   178,   542,   550,   554,   556,   559,
     459,   179,   560,   549,   250,   561,   250,   129,   563,   632,
     364,   564,   250,   244,   712,   587,   590,   589,   597,   300,
     606,   492,   631,   601,   603,    74,   610,    13,    14,   301,
     608,   509,   611,   192,   250,   614,   511,   132,   132,   132,
     616,   510,   625,   200,   633,   634,   437,   204,   247,   250,
     250,   132,   209,   210,   211,   212,   255,   510,   558,  -181,
     132,   635,  -182,   263,   636,   638,   637,   178,   132,    89,
     641,   674,   642,   640,   653,   132,   165,   644,   646,   665,
     180,   667,   512,   685,   673,   668,   684,   463,   691,   696,
     697,   247,   698,   699,    76,   373,   705,   575,   132,   702,
     666,   582,   483,   623,   613,   395,   373,   373,   624,   509,
     429,   138,   304,   695,   511,   396,   376,   479,   534,   101,
     505,   522,   607,   596,   472,   490,   694,   250,   680,   572,
     473,   592,   677,   375,   707,     0,   179,     0,     0,     0,
     250,     0,   513,     0,     0,   247,   711,   706,   510,     0,
     250,   132,     0,   132,   250,     0,     0,     0,     0,     0,
     512,   229,   229,     0,   229,   229,   192,     0,     0,     0,
     132,     0,   132,     0,   250,   250,   200,     0,   132,     0,
     204,   205,   206,   207,   208,   209,   210,   211,   212,     0,
       0,     0,     0,     0,     0,   180,     0,     0,     0,     0,
     132,     0,   178,     0,     0,     0,     0,     0,    47,     0,
       0,     0,   132,     0,     0,   132,   132,   127,     0,   152,
       0,   571,    57,    58,    12,     0,   510,   578,   581,    61,
     586,   175,     0,     0,   184,     0,   243,     0,     0,   591,
       0,   593,   513,     0,     0,     0,     0,     0,     0,   250,
       0,     0,     0,   129,     0,     0,     0,     0,   513,   244,
       0,     0,     0,     0,     0,   245,   251,   252,   180,     0,
      47,    74,     0,    13,    14,   421,     0,     0,     0,   127,
       0,     0,     0,   368,    57,    58,    12,     0,   281,   368,
     368,    61,   250,   132,     0,     0,   296,     0,   243,     0,
       0,     0,     0,   302,     0,     0,   132,     0,   132,     0,
       0,     0,     0,     0,     0,   129,   132,     0,     0,     0,
     132,   244,     0,     0,     0,     0,   311,   310,     0,     0,
       0,     0,     0,    74,   639,    13,    14,   301,     0,     0,
     132,   132,   643,     0,     0,     0,   250,     0,     0,   513,
     312,   313,   314,   315,     0,   316,   317,   318,     0,   319,
     320,   321,   322,   323,   324,   325,   326,   327,   328,   329,
     330,   331,   332,     0,   175,     0,   339,     0,   342,     0,
       0,   385,   152,   152,   353,    47,     0,     0,   261,   578,
     678,     0,     0,     0,   127,     0,     0,     0,   394,    57,
      58,    12,     0,   180,     0,     0,    61,     0,   132,     0,
       0,   368,     0,   243,   388,   132,     0,     0,   368,     0,
       0,     0,     0,     0,   132,     0,   368,   513,   422,     0,
     129,     0,     0,   190,  -300,     0,   244,     0,     0,     0,
     433,     0,     0,   439,   440,     0,     0,     0,    74,     0,
      13,    14,     0,     0,     0,     0,     0,     0,   132,     0,
     152,     0,   191,   192,     0,   193,   194,   195,   196,   197,
     152,   198,   199,   200,   201,   202,   203,   204,   205,   206,
     207,   208,   209,   210,   211,   212,     0,     0,     0,     0,
       0,     0,     0,   446,  -300,     0,     0,   175,     0,     0,
       0,     0,     0,   446,  -300,     0,     0,     0,     0,     0,
       0,     0,   132,     0,     0,   132,     0,     0,     0,     0,
       0,   433,   368,     0,     0,     0,     0,     0,   576,   368,
       0,   368,     0,     0,     0,     0,     0,     0,     0,    45,
     368,     0,   368,     0,    47,     0,     0,   183,   524,     0,
       0,   152,   152,   127,     0,     0,     0,     0,    57,    58,
      12,     0,     0,     0,     0,    61,     0,     0,   245,   544,
       0,     0,    65,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,    66,    67,     0,    68,    69,
     152,     0,    70,   132,     0,    71,     0,     0,     0,     0,
       0,     0,     0,     0,   152,    72,    73,    74,    45,    13,
      14,     0,   175,    47,     0,     0,     0,   184,     0,     0,
       0,     0,   127,     0,     0,     0,     0,    57,    58,    12,
       0,   254,     0,     0,    61,   368,     0,     0,     0,     0,
       0,    65,     0,   368,     0,     0,     0,     0,     0,     0,
     368,    45,     0,     0,    66,    67,    47,    68,    69,     0,
       0,    70,   599,   600,    71,   127,     0,     0,     0,     0,
      57,    58,    12,     0,    72,    73,    74,    61,    13,    14,
       0,     0,     0,     0,   173,   175,   618,   368,     0,     0,
     576,   368,     0,     0,     0,     0,     0,    66,    67,     0,
      68,   174,     0,     0,    70,   446,     0,    71,   334,     0,
       0,     0,   446,   621,   446,     0,     0,    72,    73,    74,
       0,    13,    14,     0,   191,   192,     0,   193,     0,   195,
     196,   197,     0,     0,   199,   200,   201,   202,   203,   204,
     205,   206,   207,   208,   209,   210,   211,   212,     0,   368,
       0,   368,    -2,    44,     0,    45,     0,     0,    46,     0,
      47,    48,    49,     0,     0,    50,   459,    51,    52,    53,
      54,    55,    56,     0,    57,    58,    12,     0,     0,    59,
      60,    61,    62,    63,    64,     0,   184,     0,    65,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,    66,    67,     0,    68,    69,     0,     0,    70,     0,
       0,    71,     0,     0,   -47,     0,     0,   688,   689,     0,
     175,    72,    73,    74,     0,    13,    14,     0,   344,   446,
      45,     0,     0,    46,  -275,    47,    48,    49,     0,  -275,
      50,     0,    51,    52,   127,    54,    55,    56,     0,    57,
      58,    12,     0,     0,    59,    60,    61,    62,    63,    64,
       0,     0,     0,    65,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,    66,    67,     0,    68,
      69,     0,     0,    70,     0,     0,    71,     0,     0,  -275,
       0,     0,     0,     0,   345,  -275,    72,    73,    74,     0,
      13,    14,   344,     0,    45,     0,     0,    46,     0,    47,
      48,    49,     0,     0,    50,     0,    51,    52,   127,    54,
      55,    56,     0,    57,    58,    12,     0,     0,    59,    60,
      61,    62,    63,    64,     0,     0,     0,    65,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
      66,    67,     0,    68,    69,     0,     0,    70,     0,    45,
      71,     0,     0,  -275,    47,     0,     0,     0,   345,  -275,
      72,    73,    74,   127,    13,    14,     0,     0,    57,    58,
      12,     0,     0,     0,     0,    61,     0,     0,     0,     0,
       0,     0,    65,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,    66,    67,     0,    68,    69,
       0,     0,    70,     0,    45,    71,     0,     0,     0,    47,
       0,     0,     0,   445,     0,    72,    73,    74,   127,    13,
      14,     0,     0,    57,    58,    12,     0,     0,     0,     0,
      61,     0,   454,     0,     0,     0,     0,   173,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
      66,    67,     0,    68,   174,     0,     0,    70,     0,    45,
      71,     0,     0,     0,    47,     0,     0,     0,     0,     0,
      72,    73,    74,   127,    13,    14,     0,     0,    57,    58,
      12,     0,   502,     0,     0,    61,     0,     0,     0,     0,
       0,     0,    65,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,    66,    67,     0,    68,    69,
       0,     0,    70,     0,    45,    71,     0,     0,     0,    47,
       0,     0,     0,     0,     0,    72,    73,    74,   127,    13,
      14,     0,     0,    57,    58,    12,     0,   503,     0,     0,
      61,     0,     0,     0,     0,     0,     0,    65,     0,     0,
       0,     0,     0,     0,     0,     0,     0,    45,     0,     0,
      66,    67,    47,    68,    69,     0,     0,    70,     0,     0,
      71,   127,     0,     0,     0,     0,    57,    58,    12,     0,
      72,    73,    74,    61,    13,    14,     0,     0,     0,     0,
      65,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,    66,    67,     0,    68,    69,     0,     0,
      70,     0,    45,    71,     0,     0,     0,    47,     0,     0,
       0,   620,     0,    72,    73,    74,   127,    13,    14,     0,
       0,    57,    58,    12,     0,     0,     0,     0,    61,     0,
       0,     0,     0,     0,     0,    65,     0,     0,     0,     0,
       0,     0,     0,     0,     0,    45,     0,     0,    66,    67,
      47,    68,    69,     0,     0,    70,     0,     0,    71,   127,
       0,     0,     0,     0,    57,    58,    12,     0,    72,    73,
      74,    61,    13,    14,     0,     0,     0,     0,   173,     0,
       0,     0,     0,     0,     0,     0,     0,     0,    45,     0,
       0,    66,    67,   303,    68,   174,     0,     0,    70,     0,
       0,    71,   127,     0,     0,     0,     0,    57,    58,    12,
       0,    72,    73,    74,    61,    13,    14,     0,     0,     0,
       0,    65,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,    66,    67,    47,    68,    69,     0,
       0,    70,     0,     0,    71,   127,     0,     0,     0,     0,
      57,    58,    12,     0,    72,    73,    74,    61,    13,    14,
       0,    47,     0,     0,   128,     0,     0,     0,     0,    47,
     127,     0,     0,     0,     0,    57,    58,    12,   127,     0,
       0,   129,    61,    57,    58,    12,     0,   130,     0,   243,
      61,     0,     0,     0,    47,     0,     0,   431,     0,    74,
       0,    13,    14,   127,     0,     0,   129,     0,    57,    58,
      12,     0,   244,     0,   129,    61,     0,     0,     0,     0,
     432,     0,   243,     0,    74,     0,    13,    14,   303,     0,
       0,     0,    74,     0,    13,    14,     0,   127,     0,   129,
       0,     0,    57,    58,    12,   506,     0,     0,     0,    61,
       0,     0,     0,   358,     0,     0,   243,    74,     0,    13,
      14,   358,   359,     0,   489,     0,     0,   360,   361,   362,
     359,     0,     0,   129,   363,   360,   361,   362,     0,   244,
       0,   469,   363,     0,     0,     0,   358,     0,     0,   364,
       0,    74,     0,    13,    14,   359,     0,     0,   365,     0,
     360,   361,   574,     0,   470,     0,   365,   363,     0,     0,
       0,     0,     0,     0,   364,     0,   367,     0,     0,    14,
     358,     0,     0,     0,   367,     0,     0,    14,     0,   359,
       0,   365,     0,     0,   360,   361,   362,     0,     0,     0,
       0,   363,     0,     0,     0,   358,     0,     0,   364,   367,
       0,    13,    14,     0,   359,     0,     0,     0,     0,   360,
     361,   362,     0,     0,     0,   365,   363,     0,     0,     0,
       0,   647,     0,   364,     0,     0,     0,     0,     0,     0,
       0,     0,     0,   367,     0,     0,    14,     0,     0,     0,
     365,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,   367,   191,
     192,    14,   193,     0,   195,   196,   197,     0,     0,   199,
     200,   201,   202,   203,   204,   205,   206,   207,   208,   209,
     210,   211,   212,     0,     0,     0,     0,     0,     0,   191,
     192,     0,   193,     0,   195,   196,   197,     0,   456,   199,
     200,   201,   202,   203,   204,   205,   206,   207,   208,   209,
     210,   211,   212,   191,   192,     0,   193,     0,   195,   196,
     197,     0,   553,   199,   200,   201,   202,   203,   204,   205,
     206,   207,   208,   209,   210,   211,   212,   191,   192,     0,
     193,     0,   195,   196,   197,     0,   703,   199,   200,   201,
     202,   203,   204,   205,   206,   207,   208,   209,   210,   211,
     212,   191,   192,     0,   193,     0,   195,   196,   197,     0,
     704,   199,   200,   201,   202,   203,   204,   205,   206,   207,
     208,   209,   210,   211,   212,   191,   192,     0,     0,     0,
     195,   196,   197,     0,     0,   199,   200,   201,   202,   203,
     204,   205,   206,   207,   208,   209,   210,   211,   212,   191,
     192,     0,     0,     0,   195,   196,   197,     0,     0,   199,
     200,   201,   202,     0,   204,   205,   206,   207,   208,   209,
     210,   211,   212,   192,     0,     0,     0,   195,   196,   197,
       0,     0,   199,   200,   201,   202,     0,   204,   205,   206,
     207,   208,   209,   210,   211,   212
};

static const yytype_int16 yycheck[] =
{
      20,   157,   402,     9,    77,   141,     6,    46,   216,    71,
      49,    31,    47,    19,   242,   158,    55,    47,   162,   219,
     343,    41,    38,   278,    30,    59,    26,   141,    28,   150,
     474,   515,   287,   147,   482,    41,   283,   284,    38,   341,
     295,   269,   336,   157,   299,    11,    46,    59,   342,    49,
      77,     3,   270,    53,   309,    55,     0,   634,    60,    20,
       3,    24,    74,    63,    64,     8,    62,   414,     1,    24,
      24,     1,    68,    75,    17,   422,     3,    77,     3,    22,
      23,    24,     3,    25,    25,   105,    29,    71,    65,    66,
      67,    68,    69,    70,     3,    72,    73,    24,   104,    24,
      62,   121,    35,     3,    59,    35,    67,   141,     7,    75,
      73,   150,     5,    12,   187,    24,    59,   156,    21,   696,
      72,   698,   606,     5,    24,     5,    59,    62,    71,    47,
      73,    74,    59,    68,    67,    60,    63,    67,    63,     3,
      50,    50,   142,   490,    54,    35,    73,    74,    73,    74,
     150,    24,    24,    71,   275,   449,   156,   451,   158,    24,
     187,   463,   162,    63,    73,    74,   413,    63,   415,    68,
      62,     7,    65,    73,    74,    62,    12,    67,     3,   407,
       3,    53,    75,    65,    66,    65,   533,   187,   222,   223,
     174,   635,     3,    75,   642,    75,   173,   174,    62,    63,
     684,    24,   646,   647,    68,    71,    62,   227,    73,    74,
     128,   129,   130,     7,   109,   215,   111,   112,    12,    66,
     226,   221,   545,   239,   142,    72,    62,    50,    62,   552,
     214,   231,    68,   151,    68,    67,    59,    24,   438,   239,
      66,   159,   242,    50,   462,   445,   593,    54,   166,   457,
      73,    74,    62,   253,    66,     7,   174,   393,    68,    61,
      12,    59,    59,   610,   611,    74,   300,   403,   303,   269,
     270,   189,    59,   303,    68,   337,   310,    59,   506,   393,
      24,   424,   426,   501,    64,    62,    73,    74,   409,   403,
      24,    93,   639,   548,    59,     9,   214,    99,    24,     8,
      63,   301,    59,    17,    75,    35,    67,    21,    17,    60,
      62,    35,   712,    22,    23,    24,    68,    31,    32,    53,
      29,    59,    72,    72,   242,    59,   244,    36,    60,    59,
      47,    35,     3,    59,    68,    35,   356,    67,     8,    73,
      74,   341,    68,   261,    53,   263,   669,    73,    74,   355,
      64,   269,    24,   337,    71,    59,    65,   391,   392,    59,
      35,    60,    71,    67,   399,    74,   366,    67,   441,   399,
     409,    35,    62,   291,    35,   603,    35,   377,   417,    24,
      75,    53,   421,    60,    59,   303,    24,    59,   306,   307,
     452,   609,    67,    62,    24,    24,   430,   397,    59,   399,
      59,    73,    74,    24,   404,    59,    67,   407,    67,   409,
     444,   128,   129,   130,    59,   399,    62,   417,   626,   337,
     620,   421,    60,    72,   424,   142,   426,    59,    73,    74,
      59,    63,     3,    67,   151,    73,    74,    62,    59,    71,
      62,   441,   159,    62,    73,    74,    62,   467,    65,   166,
      24,    65,    73,    74,    24,    66,    59,   174,    67,    24,
     466,    71,   462,   463,   499,   601,   384,    67,   452,   499,
      24,     8,   189,    65,   474,   475,    62,   477,    62,   397,
     515,   399,   482,    62,    62,   515,    60,   601,   488,   407,
      60,   491,   492,   411,    59,    63,    47,   214,    60,    73,
      74,   501,     8,    73,    74,    59,   506,    60,    73,    74,
      60,    17,    60,   431,   432,   499,    22,    23,    24,    73,
      74,    68,   230,    29,    60,   242,    68,   244,   236,   237,
      36,   515,    35,    60,   452,    60,    60,    75,    68,    60,
      75,   525,    60,   616,   261,    60,   263,    53,    60,   569,
      36,    68,   269,    59,   710,     3,    72,    62,    60,    65,
      60,    75,   568,    62,    59,    71,    60,    73,    74,    75,
      66,   606,    60,    34,   291,    60,   606,   128,   129,   130,
      62,   499,    60,    44,    60,    60,   303,    48,   506,   306,
     307,   142,    53,    54,    55,    56,   710,   515,   660,    59,
     151,    59,    59,   603,    68,    72,    62,   525,   159,   609,
      62,   631,    49,    68,    60,   166,   616,    62,    59,    68,
     337,    60,   606,    14,   630,    68,    60,    62,    72,    60,
      60,   549,    60,    60,    28,   635,    68,   637,   189,   685,
     616,   641,   642,   554,   538,   263,   646,   647,   554,   684,
     301,    49,   174,   673,   684,   263,   232,   365,   417,    28,
     397,   409,   525,   492,   358,   373,   672,   384,   641,   470,
     358,   488,   637,   231,   694,    -1,   660,    -1,    -1,    -1,
     397,    -1,   399,    -1,    -1,   603,   706,   693,   606,    -1,
     407,   242,    -1,   244,   411,    -1,    -1,    -1,    -1,    -1,
     684,   108,   109,    -1,   111,   112,    34,    -1,    -1,    -1,
     261,    -1,   263,    -1,   431,   432,    44,    -1,   269,    -1,
      48,    49,    50,    51,    52,    53,    54,    55,    56,    -1,
      -1,    -1,    -1,    -1,    -1,   452,    -1,    -1,    -1,    -1,
     291,    -1,   660,    -1,    -1,    -1,    -1,    -1,     8,    -1,
      -1,    -1,   303,    -1,    -1,   306,   307,    17,    -1,    59,
      -1,   469,    22,    23,    24,    -1,   684,   475,   476,    29,
     478,    71,    -1,    -1,    74,    -1,    36,    -1,    -1,   487,
      -1,   489,   499,    -1,    -1,    -1,    -1,    -1,    -1,   506,
      -1,    -1,    -1,    53,    -1,    -1,    -1,    -1,   515,    59,
      -1,    -1,    -1,    -1,    -1,   128,   129,   130,   525,    -1,
       8,    71,    -1,    73,    74,    75,    -1,    -1,    -1,    17,
      -1,    -1,    -1,   230,    22,    23,    24,    -1,   151,   236,
     237,    29,   549,   384,    -1,    -1,   159,    -1,    36,    -1,
      -1,    -1,    -1,   166,    -1,    -1,   397,    -1,   399,    -1,
      -1,    -1,    -1,    -1,    -1,    53,   407,    -1,    -1,    -1,
     411,    59,    -1,    -1,    -1,    -1,   189,    65,    -1,    -1,
      -1,    -1,    -1,    71,   582,    73,    74,    75,    -1,    -1,
     431,   432,   590,    -1,    -1,    -1,   603,    -1,    -1,   606,
     190,   191,   192,   193,    -1,   195,   196,   197,    -1,   199,
     200,   201,   202,   203,   204,   205,   206,   207,   208,   209,
     210,   211,   212,    -1,   214,    -1,   216,    -1,   218,    -1,
      -1,   244,   222,   223,   224,     8,    -1,    -1,    11,   637,
     638,    -1,    -1,    -1,    17,    -1,    -1,    -1,   261,    22,
      23,    24,    -1,   660,    -1,    -1,    29,    -1,   499,    -1,
      -1,   358,    -1,    36,   254,   506,    -1,    -1,   365,    -1,
      -1,    -1,    -1,    -1,   515,    -1,   373,   684,   291,    -1,
      53,    -1,    -1,     4,     5,    -1,    59,    -1,    -1,    -1,
     303,    -1,    -1,   306,   307,    -1,    -1,    -1,    71,    -1,
      73,    74,    -1,    -1,    -1,    -1,    -1,    -1,   549,    -1,
     300,    -1,    33,    34,    -1,    36,    37,    38,    39,    40,
     310,    42,    43,    44,    45,    46,    47,    48,    49,    50,
      51,    52,    53,    54,    55,    56,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,   333,    65,    -1,    -1,   337,    -1,    -1,
      -1,    -1,    -1,   343,    75,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,   603,    -1,    -1,   606,    -1,    -1,    -1,    -1,
      -1,   384,   469,    -1,    -1,    -1,    -1,    -1,   475,   476,
      -1,   478,    -1,    -1,    -1,    -1,    -1,    -1,    -1,     3,
     487,    -1,   489,    -1,     8,    -1,    -1,    11,   411,    -1,
      -1,   391,   392,    17,    -1,    -1,    -1,    -1,    22,    23,
      24,    -1,    -1,    -1,    -1,    29,    -1,    -1,   431,   432,
      -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    49,    50,    -1,    52,    53,
     430,    -1,    56,   684,    -1,    59,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,   444,    69,    70,    71,     3,    73,
      74,    -1,   452,     8,    -1,    -1,    -1,   457,    -1,    -1,
      -1,    -1,    17,    -1,    -1,    -1,    -1,    22,    23,    24,
      -1,    26,    -1,    -1,    29,   582,    -1,    -1,    -1,    -1,
      -1,    36,    -1,   590,    -1,    -1,    -1,    -1,    -1,    -1,
     597,     3,    -1,    -1,    49,    50,     8,    52,    53,    -1,
      -1,    56,   502,   503,    59,    17,    -1,    -1,    -1,    -1,
      22,    23,    24,    -1,    69,    70,    71,    29,    73,    74,
      -1,    -1,    -1,    -1,    36,   525,   549,   634,    -1,    -1,
     637,   638,    -1,    -1,    -1,    -1,    -1,    49,    50,    -1,
      52,    53,    -1,    -1,    56,   545,    -1,    59,    60,    -1,
      -1,    -1,   552,   553,   554,    -1,    -1,    69,    70,    71,
      -1,    73,    74,    -1,    33,    34,    -1,    36,    -1,    38,
      39,    40,    -1,    -1,    43,    44,    45,    46,    47,    48,
      49,    50,    51,    52,    53,    54,    55,    56,    -1,   696,
      -1,   698,     0,     1,    -1,     3,    -1,    -1,     6,    -1,
       8,     9,    10,    -1,    -1,    13,    75,    15,    16,    17,
      18,    19,    20,    -1,    22,    23,    24,    -1,    -1,    27,
      28,    29,    30,    31,    32,    -1,   626,    -1,    36,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    49,    50,    -1,    52,    53,    -1,    -1,    56,    -1,
      -1,    59,    -1,    -1,    62,    -1,    -1,   657,   658,    -1,
     660,    69,    70,    71,    -1,    73,    74,    -1,     1,   669,
       3,    -1,    -1,     6,     7,     8,     9,    10,    -1,    12,
      13,    -1,    15,    16,    17,    18,    19,    20,    -1,    22,
      23,    24,    -1,    -1,    27,    28,    29,    30,    31,    32,
      -1,    -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    49,    50,    -1,    52,
      53,    -1,    -1,    56,    -1,    -1,    59,    -1,    -1,    62,
      -1,    -1,    -1,    -1,    67,    68,    69,    70,    71,    -1,
      73,    74,     1,    -1,     3,    -1,    -1,     6,    -1,     8,
       9,    10,    -1,    -1,    13,    -1,    15,    16,    17,    18,
      19,    20,    -1,    22,    23,    24,    -1,    -1,    27,    28,
      29,    30,    31,    32,    -1,    -1,    -1,    36,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      49,    50,    -1,    52,    53,    -1,    -1,    56,    -1,     3,
      59,    -1,    -1,    62,     8,    -1,    -1,    -1,    67,    68,
      69,    70,    71,    17,    73,    74,    -1,    -1,    22,    23,
      24,    -1,    -1,    -1,    -1,    29,    -1,    -1,    -1,    -1,
      -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    49,    50,    -1,    52,    53,
      -1,    -1,    56,    -1,     3,    59,    -1,    -1,    -1,     8,
      -1,    -1,    -1,    67,    -1,    69,    70,    71,    17,    73,
      74,    -1,    -1,    22,    23,    24,    -1,    -1,    -1,    -1,
      29,    -1,    31,    -1,    -1,    -1,    -1,    36,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      49,    50,    -1,    52,    53,    -1,    -1,    56,    -1,     3,
      59,    -1,    -1,    -1,     8,    -1,    -1,    -1,    -1,    -1,
      69,    70,    71,    17,    73,    74,    -1,    -1,    22,    23,
      24,    -1,    26,    -1,    -1,    29,    -1,    -1,    -1,    -1,
      -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    49,    50,    -1,    52,    53,
      -1,    -1,    56,    -1,     3,    59,    -1,    -1,    -1,     8,
      -1,    -1,    -1,    -1,    -1,    69,    70,    71,    17,    73,
      74,    -1,    -1,    22,    23,    24,    -1,    26,    -1,    -1,
      29,    -1,    -1,    -1,    -1,    -1,    -1,    36,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,     3,    -1,    -1,
      49,    50,     8,    52,    53,    -1,    -1,    56,    -1,    -1,
      59,    17,    -1,    -1,    -1,    -1,    22,    23,    24,    -1,
      69,    70,    71,    29,    73,    74,    -1,    -1,    -1,    -1,
      36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    49,    50,    -1,    52,    53,    -1,    -1,
      56,    -1,     3,    59,    -1,    -1,    -1,     8,    -1,    -1,
      -1,    67,    -1,    69,    70,    71,    17,    73,    74,    -1,
      -1,    22,    23,    24,    -1,    -1,    -1,    -1,    29,    -1,
      -1,    -1,    -1,    -1,    -1,    36,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,     3,    -1,    -1,    49,    50,
       8,    52,    53,    -1,    -1,    56,    -1,    -1,    59,    17,
      -1,    -1,    -1,    -1,    22,    23,    24,    -1,    69,    70,
      71,    29,    73,    74,    -1,    -1,    -1,    -1,    36,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,     3,    -1,
      -1,    49,    50,     8,    52,    53,    -1,    -1,    56,    -1,
      -1,    59,    17,    -1,    -1,    -1,    -1,    22,    23,    24,
      -1,    69,    70,    71,    29,    73,    74,    -1,    -1,    -1,
      -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    49,    50,     8,    52,    53,    -1,
      -1,    56,    -1,    -1,    59,    17,    -1,    -1,    -1,    -1,
      22,    23,    24,    -1,    69,    70,    71,    29,    73,    74,
      -1,     8,    -1,    -1,    36,    -1,    -1,    -1,    -1,     8,
      17,    -1,    -1,    -1,    -1,    22,    23,    24,    17,    -1,
      -1,    53,    29,    22,    23,    24,    -1,    59,    -1,    36,
      29,    -1,    -1,    -1,     8,    -1,    -1,    36,    -1,    71,
      -1,    73,    74,    17,    -1,    -1,    53,    -1,    22,    23,
      24,    -1,    59,    -1,    53,    29,    -1,    -1,    -1,    -1,
      59,    -1,    36,    -1,    71,    -1,    73,    74,     8,    -1,
      -1,    -1,    71,    -1,    73,    74,    -1,    17,    -1,    53,
      -1,    -1,    22,    23,    24,    59,    -1,    -1,    -1,    29,
      -1,    -1,    -1,     8,    -1,    -1,    36,    71,    -1,    73,
      74,     8,    17,    -1,    11,    -1,    -1,    22,    23,    24,
      17,    -1,    -1,    53,    29,    22,    23,    24,    -1,    59,
      -1,    36,    29,    -1,    -1,    -1,     8,    -1,    -1,    36,
      -1,    71,    -1,    73,    74,    17,    -1,    -1,    53,    -1,
      22,    23,    24,    -1,    59,    -1,    53,    29,    -1,    -1,
      -1,    -1,    -1,    -1,    36,    -1,    71,    -1,    -1,    74,
       8,    -1,    -1,    -1,    71,    -1,    -1,    74,    -1,    17,
      -1,    53,    -1,    -1,    22,    23,    24,    -1,    -1,    -1,
      -1,    29,    -1,    -1,    -1,     8,    -1,    -1,    36,    71,
      -1,    73,    74,    -1,    17,    -1,    -1,    -1,    -1,    22,
      23,    24,    -1,    -1,    -1,    53,    29,    -1,    -1,    -1,
      -1,    59,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    71,    -1,    -1,    74,    -1,    -1,    -1,
      53,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    71,    33,
      34,    74,    36,    -1,    38,    39,    40,    -1,    -1,    43,
      44,    45,    46,    47,    48,    49,    50,    51,    52,    53,
      54,    55,    56,    -1,    -1,    -1,    -1,    -1,    -1,    33,
      34,    -1,    36,    -1,    38,    39,    40,    -1,    72,    43,
      44,    45,    46,    47,    48,    49,    50,    51,    52,    53,
      54,    55,    56,    33,    34,    -1,    36,    -1,    38,    39,
      40,    -1,    66,    43,    44,    45,    46,    47,    48,    49,
      50,    51,    52,    53,    54,    55,    56,    33,    34,    -1,
      36,    -1,    38,    39,    40,    -1,    66,    43,    44,    45,
      46,    47,    48,    49,    50,    51,    52,    53,    54,    55,
      56,    33,    34,    -1,    36,    -1,    38,    39,    40,    -1,
      66,    43,    44,    45,    46,    47,    48,    49,    50,    51,
      52,    53,    54,    55,    56,    33,    34,    -1,    -1,    -1,
      38,    39,    40,    -1,    -1,    43,    44,    45,    46,    47,
      48,    49,    50,    51,    52,    53,    54,    55,    56,    33,
      34,    -1,    -1,    -1,    38,    39,    40,    -1,    -1,    43,
      44,    45,    46,    -1,    48,    49,    50,    51,    52,    53,
      54,    55,    56,    34,    -1,    -1,    -1,    38,    39,    40,
      -1,    -1,    43,    44,    45,    46,    -1,    48,    49,    50,
      51,    52,    53,    54,    55,    56
};

/* YYSTOS[STATE-NUM] -- The (internal number of the) accessing
   symbol of state STATE-NUM.  */
static const yytype_uint8 yystos[] =
{
       0,    77,    78,    80,    81,     0,    25,    79,   100,   101,
      25,   109,    24,    73,    74,   164,   165,   104,   102,   103,
     109,    24,   111,   112,     3,    62,    21,   105,   189,    90,
      91,   109,   111,    24,   110,   237,    63,     3,    59,    63,
     106,   108,   164,    62,     1,     3,     6,     8,     9,    10,
      13,    15,    16,    17,    18,    19,    20,    22,    23,    27,
      28,    29,    30,    31,    32,    36,    49,    50,    52,    53,
      56,    59,    69,    70,    71,   113,   114,   115,   121,   133,
     136,   144,   147,   149,   150,   151,   152,   157,   161,   164,
     166,   167,   172,   173,   176,   179,   180,   181,   184,   187,
     188,   204,   209,    82,    83,   109,   111,    62,     9,    17,
      21,    31,    32,    64,   222,    24,    73,    60,   106,   107,
       3,   109,   111,     3,   161,   163,   164,    17,    36,    53,
      59,   164,   166,   171,   175,   176,   177,   184,   163,   151,
     157,   134,    59,   164,   182,   151,   161,   137,    35,    67,
     160,    71,   149,   209,   216,   148,   160,   145,    59,   119,
     120,   164,    59,   116,   162,   164,   208,   150,   150,   150,
     150,   150,   150,    36,    53,   149,   158,   170,   176,   178,
     184,   150,   150,    11,   149,   215,    62,    59,   117,   208,
       4,    33,    34,    36,    37,    38,    39,    40,    42,    43,
      44,    45,    46,    47,    48,    49,    50,    51,    52,    53,
      54,    55,    56,    67,    59,    63,    71,    66,    59,   160,
       1,   160,     5,    65,    75,    92,    93,   109,   111,   165,
     223,    59,   183,   223,    24,   223,   224,   223,    64,    62,
     213,   111,    59,    36,    59,   169,   175,   176,   177,   178,
     184,   169,   169,    63,    26,   121,   130,   131,   132,   209,
     217,    11,   159,   164,   168,   169,   200,   201,   202,    59,
      67,   185,   135,   217,    24,    59,    68,   161,   194,   196,
     198,   169,    35,    53,    59,    68,   161,   193,   195,   196,
     197,   207,   135,    60,   120,   192,   169,    60,   116,   190,
      65,    75,   169,     8,   170,    60,    72,    72,    60,   117,
      65,   169,   149,   149,   149,   149,   149,   149,   149,   149,
     149,   149,   149,   149,   149,   149,   149,   149,   149,   149,
     149,   149,   149,   153,    60,   158,   210,    59,   164,   149,
     215,   205,   149,   153,     1,    67,   114,   123,   203,   204,
     206,   209,   209,   149,    94,    95,   109,   111,     8,    17,
      22,    23,    24,    29,    36,    53,    65,    71,   165,   225,
     227,   228,   229,   164,   230,   238,   185,    59,     3,   225,
     225,   106,    60,   202,     8,   169,    60,   164,   149,    35,
     128,     5,    65,    62,   169,   159,   168,    75,   214,    60,
     202,   206,   138,    62,    63,    24,   196,    59,   199,    62,
     213,    72,   127,    59,   197,    53,   197,    62,   213,     3,
     221,    75,   169,   146,    62,   213,    62,   213,   209,   162,
      65,    36,    59,   169,   175,   176,   177,   184,    67,   169,
     169,    62,   213,   209,    65,    67,   149,   154,   155,   211,
     212,    11,    75,   214,    31,   158,    72,    66,   203,    75,
     214,   212,   124,    62,    68,    98,    99,   109,   111,    36,
      59,   226,   227,   229,    59,    67,    71,    67,     8,   225,
       3,    50,    59,   164,   235,   236,     3,    72,    65,    11,
     225,    60,    75,    62,   218,   238,    62,    62,    62,    60,
      60,   129,    26,    26,   217,   200,    59,   164,   174,   175,
     176,   177,   178,   184,   186,    60,    68,   128,   217,   164,
      60,   202,   198,    68,   169,     7,    12,    68,   122,   125,
     197,   221,   197,    60,   195,    68,   161,   221,    35,   120,
      60,   116,    60,   209,   169,   153,   117,   118,   191,   208,
      60,   209,   153,    66,    75,   214,    68,   214,   158,    60,
      60,    60,   215,    60,    68,   206,   203,    96,    97,   109,
     111,   225,   228,   218,    24,   164,   165,   220,   225,   232,
     240,   225,   164,   219,   231,   239,   225,     3,   235,    62,
      72,   225,   236,   225,   221,   164,   230,    60,   206,   149,
     149,    62,   202,    59,   186,   139,    60,   210,    66,   126,
      60,    60,   221,   127,    60,   212,    62,   213,   169,   212,
      67,   149,   156,   154,   155,    60,    66,    72,    68,    84,
      85,   109,   111,    60,    60,    59,    68,    62,    72,   225,
      68,    62,    49,   225,    62,   221,    59,    59,   225,   233,
     234,    68,   217,    60,   202,   142,   186,     5,    65,    66,
      75,   206,   221,   221,    68,    68,   118,    60,    68,   153,
     215,    86,    87,   109,   111,   233,   218,   232,   225,   221,
     231,   235,   218,   218,    60,    14,   140,   143,   149,   149,
     212,    72,    88,    89,   109,   111,    60,    60,    60,    60,
     186,    20,   123,    66,    66,    68,   109,   111,   233,   233,
     141,   111,   135,   128
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
#line 128 "go.y"
    {
		xtop = concat(xtop, (yyvsp[(4) - (4)].list));
	}
    break;

  case 3:
#line 145 "go.y"
    {
	}
    break;

  case 4:
#line 150 "go.y"
    {
		prevlineno = lineno;
		yyerror("package statement must be first");
		errorexit();
	}
    break;

  case 5:
#line 156 "go.y"
    {
		mkpackage((yyvsp[(2) - (3)].sym)->name);
	}
    break;

  case 6:
#line 166 "go.y"
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

  case 7:
#line 178 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 8:
#line 183 "go.y"
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

  case 9:
#line 195 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 10:
#line 201 "go.y"
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

  case 11:
#line 213 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 12:
#line 219 "go.y"
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

  case 13:
#line 231 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 14:
#line 237 "go.y"
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

  case 15:
#line 249 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 16:
#line 255 "go.y"
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

  case 17:
#line 267 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 18:
#line 273 "go.y"
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

  case 19:
#line 285 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 20:
#line 290 "go.y"
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

  case 21:
#line 302 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 22:
#line 308 "go.y"
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

  case 23:
#line 320 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 24:
#line 326 "go.y"
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

  case 25:
#line 338 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 26:
#line 344 "go.y"
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

  case 27:
#line 356 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 28:
#line 367 "go.y"
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

  case 29:
#line 378 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 35:
#line 393 "go.y"
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

  case 36:
#line 430 "go.y"
    {
		// When an invalid import path is passed to importfile,
		// it calls yyerror and then sets up a fake import with
		// no package statement. This allows us to test more
		// than one invalid import statement in a single file.
		if(nerrors == 0)
			fatal("phase error in import");
	}
    break;

  case 39:
#line 445 "go.y"
    {
		// import with original name
		(yyval.i) = parserline();
		importmyname = S;
		importfile(&(yyvsp[(1) - (1)].val), (yyval.i));
	}
    break;

  case 40:
#line 452 "go.y"
    {
		// import with given name
		(yyval.i) = parserline();
		importmyname = (yyvsp[(1) - (2)].sym);
		importfile(&(yyvsp[(2) - (2)].val), (yyval.i));
	}
    break;

  case 41:
#line 459 "go.y"
    {
		// import into my name space
		(yyval.i) = parserline();
		importmyname = lookup(".");
		importfile(&(yyvsp[(2) - (2)].val), (yyval.i));
	}
    break;

  case 42:
#line 468 "go.y"
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

  case 44:
#line 483 "go.y"
    {
		if(strcmp((yyvsp[(1) - (1)].sym)->name, "safe") == 0)
			curio.importsafe = 1;
	}
    break;

  case 45:
#line 489 "go.y"
    {
		defercheckwidth();
	}
    break;

  case 46:
#line 493 "go.y"
    {
		resumecheckwidth();
		unimportfile();
	}
    break;

  case 47:
#line 502 "go.y"
    {
		yyerror("empty top-level declaration");
		(yyval.list) = nil;
	}
    break;

  case 49:
#line 508 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 50:
#line 512 "go.y"
    {
		yyerror("non-declaration statement outside function body");
		(yyval.list) = nil;
	}
    break;

  case 51:
#line 517 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 52:
#line 523 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (2)].list);
	}
    break;

  case 53:
#line 527 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (5)].list);
	}
    break;

  case 54:
#line 531 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 55:
#line 535 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (2)].list);
		iota = -100000;
		lastconst = nil;
	}
    break;

  case 56:
#line 541 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (5)].list);
		iota = -100000;
		lastconst = nil;
	}
    break;

  case 57:
#line 547 "go.y"
    {
		(yyval.list) = concat((yyvsp[(3) - (7)].list), (yyvsp[(5) - (7)].list));
		iota = -100000;
		lastconst = nil;
	}
    break;

  case 58:
#line 553 "go.y"
    {
		(yyval.list) = nil;
		iota = -100000;
	}
    break;

  case 59:
#line 558 "go.y"
    {
		(yyval.list) = list1((yyvsp[(2) - (2)].node));
	}
    break;

  case 60:
#line 562 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (5)].list);
	}
    break;

  case 61:
#line 566 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 62:
#line 572 "go.y"
    {
		iota = 0;
	}
    break;

  case 63:
#line 578 "go.y"
    {
		(yyval.list) = variter((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].node), nil);
	}
    break;

  case 64:
#line 582 "go.y"
    {
		(yyval.list) = variter((yyvsp[(1) - (4)].list), (yyvsp[(2) - (4)].node), (yyvsp[(4) - (4)].list));
	}
    break;

  case 65:
#line 586 "go.y"
    {
		(yyval.list) = variter((yyvsp[(1) - (3)].list), nil, (yyvsp[(3) - (3)].list));
	}
    break;

  case 66:
#line 592 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (4)].list), (yyvsp[(2) - (4)].node), (yyvsp[(4) - (4)].list));
	}
    break;

  case 67:
#line 596 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (3)].list), N, (yyvsp[(3) - (3)].list));
	}
    break;

  case 69:
#line 603 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].node), nil);
	}
    break;

  case 70:
#line 607 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (1)].list), N, nil);
	}
    break;

  case 71:
#line 613 "go.y"
    {
		// different from dclname because the name
		// becomes visible right here, not at the end
		// of the declaration.
		(yyval.node) = typedcl0((yyvsp[(1) - (1)].sym));
	}
    break;

  case 72:
#line 622 "go.y"
    {
		(yyval.node) = typedcl1((yyvsp[(1) - (2)].node), (yyvsp[(2) - (2)].node), 1);
	}
    break;

  case 73:
#line 628 "go.y"
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

  case 74:
#line 646 "go.y"
    {
		(yyval.node) = nod(OASOP, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
		(yyval.node)->etype = (yyvsp[(2) - (3)].i);			// rathole to pass opcode
	}
    break;

  case 75:
#line 651 "go.y"
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

  case 76:
#line 663 "go.y"
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

  case 77:
#line 679 "go.y"
    {
		(yyval.node) = nod(OASOP, (yyvsp[(1) - (2)].node), nodintconst(1));
		(yyval.node)->implicit = 1;
		(yyval.node)->etype = OADD;
	}
    break;

  case 78:
#line 685 "go.y"
    {
		(yyval.node) = nod(OASOP, (yyvsp[(1) - (2)].node), nodintconst(1));
		(yyval.node)->implicit = 1;
		(yyval.node)->etype = OSUB;
	}
    break;

  case 79:
#line 693 "go.y"
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

  case 80:
#line 713 "go.y"
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

  case 81:
#line 731 "go.y"
    {
		// will be converted to OCASE
		// right will point to next case
		// done in casebody()
		markdcl();
		(yyval.node) = nod(OXCASE, N, N);
		(yyval.node)->list = list1(colas((yyvsp[(2) - (5)].list), list1((yyvsp[(4) - (5)].node)), (yyvsp[(3) - (5)].i)));
	}
    break;

  case 82:
#line 740 "go.y"
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

  case 83:
#line 758 "go.y"
    {
		markdcl();
	}
    break;

  case 84:
#line 762 "go.y"
    {
		if((yyvsp[(3) - (4)].list) == nil)
			(yyval.node) = nod(OEMPTY, N, N);
		else
			(yyval.node) = liststmt((yyvsp[(3) - (4)].list));
		popdcl();
	}
    break;

  case 85:
#line 772 "go.y"
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

  case 86:
#line 783 "go.y"
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

  case 87:
#line 803 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 88:
#line 807 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].node));
	}
    break;

  case 89:
#line 813 "go.y"
    {
		markdcl();
	}
    break;

  case 90:
#line 817 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (4)].list);
		popdcl();
	}
    break;

  case 91:
#line 824 "go.y"
    {
		(yyval.node) = nod(ORANGE, N, (yyvsp[(4) - (4)].node));
		(yyval.node)->list = (yyvsp[(1) - (4)].list);
		(yyval.node)->etype = 0;	// := flag
	}
    break;

  case 92:
#line 830 "go.y"
    {
		(yyval.node) = nod(ORANGE, N, (yyvsp[(4) - (4)].node));
		(yyval.node)->list = (yyvsp[(1) - (4)].list);
		(yyval.node)->colas = 1;
		colasdefn((yyvsp[(1) - (4)].list), (yyval.node));
	}
    break;

  case 93:
#line 837 "go.y"
    {
		(yyval.node) = nod(ORANGE, N, (yyvsp[(2) - (2)].node));
		(yyval.node)->etype = 0; // := flag
	}
    break;

  case 94:
#line 844 "go.y"
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

  case 95:
#line 855 "go.y"
    {
		// normal test
		(yyval.node) = nod(OFOR, N, N);
		(yyval.node)->ntest = (yyvsp[(1) - (1)].node);
	}
    break;

  case 97:
#line 864 "go.y"
    {
		(yyval.node) = (yyvsp[(1) - (2)].node);
		(yyval.node)->nbody = concat((yyval.node)->nbody, (yyvsp[(2) - (2)].list));
	}
    break;

  case 98:
#line 871 "go.y"
    {
		markdcl();
	}
    break;

  case 99:
#line 875 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (3)].node);
		popdcl();
	}
    break;

  case 100:
#line 882 "go.y"
    {
		// test
		(yyval.node) = nod(OIF, N, N);
		(yyval.node)->ntest = (yyvsp[(1) - (1)].node);
	}
    break;

  case 101:
#line 888 "go.y"
    {
		// init ; test
		(yyval.node) = nod(OIF, N, N);
		if((yyvsp[(1) - (3)].node) != N)
			(yyval.node)->ninit = list1((yyvsp[(1) - (3)].node));
		(yyval.node)->ntest = (yyvsp[(3) - (3)].node);
	}
    break;

  case 102:
#line 899 "go.y"
    {
		markdcl();
	}
    break;

  case 103:
#line 903 "go.y"
    {
		if((yyvsp[(3) - (3)].node)->ntest == N)
			yyerror("missing condition in if statement");
	}
    break;

  case 104:
#line 908 "go.y"
    {
		(yyvsp[(3) - (5)].node)->nbody = (yyvsp[(5) - (5)].list);
	}
    break;

  case 105:
#line 912 "go.y"
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

  case 106:
#line 929 "go.y"
    {
		markdcl();
	}
    break;

  case 107:
#line 933 "go.y"
    {
		if((yyvsp[(4) - (5)].node)->ntest == N)
			yyerror("missing condition in if statement");
		(yyvsp[(4) - (5)].node)->nbody = (yyvsp[(5) - (5)].list);
		(yyval.list) = list1((yyvsp[(4) - (5)].node));
	}
    break;

  case 108:
#line 941 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 109:
#line 945 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].list));
	}
    break;

  case 110:
#line 950 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 111:
#line 954 "go.y"
    {
		NodeList *node;
		
		node = mal(sizeof *node);
		node->n = (yyvsp[(2) - (2)].node);
		node->end = node;
		(yyval.list) = node;
	}
    break;

  case 112:
#line 965 "go.y"
    {
		markdcl();
	}
    break;

  case 113:
#line 969 "go.y"
    {
		Node *n;
		n = (yyvsp[(3) - (3)].node)->ntest;
		if(n != N && n->op != OTYPESW)
			n = N;
		typesw = nod(OXXX, typesw, n);
	}
    break;

  case 114:
#line 977 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (7)].node);
		(yyval.node)->op = OSWITCH;
		(yyval.node)->list = (yyvsp[(6) - (7)].list);
		typesw = typesw->left;
		popdcl();
	}
    break;

  case 115:
#line 987 "go.y"
    {
		typesw = nod(OXXX, typesw, N);
	}
    break;

  case 116:
#line 991 "go.y"
    {
		(yyval.node) = nod(OSELECT, N, N);
		(yyval.node)->lineno = typesw->lineno;
		(yyval.node)->list = (yyvsp[(4) - (5)].list);
		typesw = typesw->left;
	}
    break;

  case 118:
#line 1004 "go.y"
    {
		(yyval.node) = nod(OOROR, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 119:
#line 1008 "go.y"
    {
		(yyval.node) = nod(OANDAND, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 120:
#line 1012 "go.y"
    {
		(yyval.node) = nod(OEQ, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 121:
#line 1016 "go.y"
    {
		(yyval.node) = nod(ONE, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 122:
#line 1020 "go.y"
    {
		(yyval.node) = nod(OLT, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 123:
#line 1024 "go.y"
    {
		(yyval.node) = nod(OLE, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 124:
#line 1028 "go.y"
    {
		(yyval.node) = nod(OGE, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 125:
#line 1032 "go.y"
    {
		(yyval.node) = nod(OGT, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 126:
#line 1036 "go.y"
    {
		(yyval.node) = nod(OADD, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 127:
#line 1040 "go.y"
    {
		(yyval.node) = nod(OSUB, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 128:
#line 1044 "go.y"
    {
		(yyval.node) = nod(OOR, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 129:
#line 1048 "go.y"
    {
		(yyval.node) = nod(OXOR, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 130:
#line 1052 "go.y"
    {
		(yyval.node) = nod(OMUL, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 131:
#line 1056 "go.y"
    {
		(yyval.node) = nod(ODIV, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 132:
#line 1060 "go.y"
    {
		(yyval.node) = nod(OMOD, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 133:
#line 1064 "go.y"
    {
		(yyval.node) = nod(OAND, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 134:
#line 1068 "go.y"
    {
		(yyval.node) = nod(OANDNOT, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 135:
#line 1072 "go.y"
    {
		(yyval.node) = nod(OLSH, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 136:
#line 1076 "go.y"
    {
		(yyval.node) = nod(ORSH, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 137:
#line 1081 "go.y"
    {
		(yyval.node) = nod(OSEND, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 139:
#line 1088 "go.y"
    {
		(yyval.node) = nod(OIND, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 140:
#line 1092 "go.y"
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

  case 141:
#line 1103 "go.y"
    {
		(yyval.node) = nod(OPLUS, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 142:
#line 1107 "go.y"
    {
		(yyval.node) = nod(OMINUS, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 143:
#line 1111 "go.y"
    {
		(yyval.node) = nod(ONOT, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 144:
#line 1115 "go.y"
    {
		yyerror("the bitwise complement operator is ^");
		(yyval.node) = nod(OCOM, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 145:
#line 1120 "go.y"
    {
		(yyval.node) = nod(OCOM, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 146:
#line 1124 "go.y"
    {
		(yyval.node) = nod(ORECV, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 147:
#line 1134 "go.y"
    {
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (3)].node), N);
	}
    break;

  case 148:
#line 1138 "go.y"
    {
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (5)].node), N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
	}
    break;

  case 149:
#line 1143 "go.y"
    {
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (6)].node), N);
		(yyval.node)->list = (yyvsp[(3) - (6)].list);
		(yyval.node)->isddd = 1;
	}
    break;

  case 150:
#line 1151 "go.y"
    {
		(yyval.node) = nodlit((yyvsp[(1) - (1)].val));
	}
    break;

  case 152:
#line 1156 "go.y"
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

  case 153:
#line 1167 "go.y"
    {
		(yyval.node) = nod(ODOTTYPE, (yyvsp[(1) - (5)].node), (yyvsp[(4) - (5)].node));
	}
    break;

  case 154:
#line 1171 "go.y"
    {
		(yyval.node) = nod(OTYPESW, N, (yyvsp[(1) - (5)].node));
	}
    break;

  case 155:
#line 1175 "go.y"
    {
		(yyval.node) = nod(OINDEX, (yyvsp[(1) - (4)].node), (yyvsp[(3) - (4)].node));
	}
    break;

  case 156:
#line 1179 "go.y"
    {
		(yyval.node) = nod(OSLICE, (yyvsp[(1) - (6)].node), nod(OKEY, (yyvsp[(3) - (6)].node), (yyvsp[(5) - (6)].node)));
	}
    break;

  case 157:
#line 1183 "go.y"
    {
		if((yyvsp[(5) - (8)].node) == N)
			yyerror("middle index required in 3-index slice");
		if((yyvsp[(7) - (8)].node) == N)
			yyerror("final index required in 3-index slice");
		(yyval.node) = nod(OSLICE3, (yyvsp[(1) - (8)].node), nod(OKEY, (yyvsp[(3) - (8)].node), nod(OKEY, (yyvsp[(5) - (8)].node), (yyvsp[(7) - (8)].node))));
	}
    break;

  case 159:
#line 1192 "go.y"
    {
		// conversion
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (5)].node), N);
		(yyval.node)->list = list1((yyvsp[(3) - (5)].node));
	}
    break;

  case 160:
#line 1198 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (5)].node);
		(yyval.node)->right = (yyvsp[(1) - (5)].node);
		(yyval.node)->list = (yyvsp[(4) - (5)].list);
		fixlbrace((yyvsp[(2) - (5)].i));
	}
    break;

  case 161:
#line 1205 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (5)].node);
		(yyval.node)->right = (yyvsp[(1) - (5)].node);
		(yyval.node)->list = (yyvsp[(4) - (5)].list);
	}
    break;

  case 162:
#line 1211 "go.y"
    {
		yyerror("cannot parenthesize type in composite literal");
		(yyval.node) = (yyvsp[(5) - (7)].node);
		(yyval.node)->right = (yyvsp[(2) - (7)].node);
		(yyval.node)->list = (yyvsp[(6) - (7)].list);
	}
    break;

  case 164:
#line 1220 "go.y"
    {
		// composite expression.
		// make node early so we get the right line number.
		(yyval.node) = nod(OCOMPLIT, N, N);
	}
    break;

  case 165:
#line 1228 "go.y"
    {
		(yyval.node) = nod(OKEY, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 166:
#line 1234 "go.y"
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

  case 167:
#line 1251 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (4)].node);
		(yyval.node)->list = (yyvsp[(3) - (4)].list);
	}
    break;

  case 169:
#line 1259 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (4)].node);
		(yyval.node)->list = (yyvsp[(3) - (4)].list);
	}
    break;

  case 171:
#line 1267 "go.y"
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

  case 175:
#line 1293 "go.y"
    {
		(yyval.i) = LBODY;
	}
    break;

  case 176:
#line 1297 "go.y"
    {
		(yyval.i) = '{';
	}
    break;

  case 177:
#line 1308 "go.y"
    {
		if((yyvsp[(1) - (1)].sym) == S)
			(yyval.node) = N;
		else
			(yyval.node) = newname((yyvsp[(1) - (1)].sym));
	}
    break;

  case 178:
#line 1317 "go.y"
    {
		(yyval.node) = dclname((yyvsp[(1) - (1)].sym));
	}
    break;

  case 179:
#line 1322 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 181:
#line 1329 "go.y"
    {
		(yyval.sym) = (yyvsp[(1) - (1)].sym);
		// during imports, unqualified non-exported identifiers are from builtinpkg
		if(importpkg != nil && !exportname((yyvsp[(1) - (1)].sym)->name))
			(yyval.sym) = pkglookup((yyvsp[(1) - (1)].sym)->name, builtinpkg);
	}
    break;

  case 183:
#line 1337 "go.y"
    {
		(yyval.sym) = S;
	}
    break;

  case 184:
#line 1343 "go.y"
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

  case 185:
#line 1356 "go.y"
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

  case 186:
#line 1371 "go.y"
    {
		(yyval.node) = oldname((yyvsp[(1) - (1)].sym));
		if((yyval.node)->pack != N)
			(yyval.node)->pack->used = 1;
	}
    break;

  case 188:
#line 1391 "go.y"
    {
		yyerror("final argument in variadic function missing type");
		(yyval.node) = nod(ODDD, typenod(typ(TINTER)), N);
	}
    break;

  case 189:
#line 1396 "go.y"
    {
		(yyval.node) = nod(ODDD, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 195:
#line 1407 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
	}
    break;

  case 199:
#line 1416 "go.y"
    {
		(yyval.node) = nod(OIND, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 204:
#line 1426 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
	}
    break;

  case 214:
#line 1447 "go.y"
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

  case 215:
#line 1460 "go.y"
    {
		(yyval.node) = nod(OTARRAY, (yyvsp[(2) - (4)].node), (yyvsp[(4) - (4)].node));
	}
    break;

  case 216:
#line 1464 "go.y"
    {
		// array literal of nelem
		(yyval.node) = nod(OTARRAY, nod(ODDD, N, N), (yyvsp[(4) - (4)].node));
	}
    break;

  case 217:
#line 1469 "go.y"
    {
		(yyval.node) = nod(OTCHAN, (yyvsp[(2) - (2)].node), N);
		(yyval.node)->etype = Cboth;
	}
    break;

  case 218:
#line 1474 "go.y"
    {
		(yyval.node) = nod(OTCHAN, (yyvsp[(3) - (3)].node), N);
		(yyval.node)->etype = Csend;
	}
    break;

  case 219:
#line 1479 "go.y"
    {
		(yyval.node) = nod(OTMAP, (yyvsp[(3) - (5)].node), (yyvsp[(5) - (5)].node));
	}
    break;

  case 222:
#line 1487 "go.y"
    {
		(yyval.node) = nod(OIND, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 223:
#line 1493 "go.y"
    {
		(yyval.node) = nod(OTCHAN, (yyvsp[(3) - (3)].node), N);
		(yyval.node)->etype = Crecv;
	}
    break;

  case 224:
#line 1500 "go.y"
    {
		(yyval.node) = nod(OTSTRUCT, N, N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
		fixlbrace((yyvsp[(2) - (5)].i));
	}
    break;

  case 225:
#line 1506 "go.y"
    {
		(yyval.node) = nod(OTSTRUCT, N, N);
		fixlbrace((yyvsp[(2) - (3)].i));
	}
    break;

  case 226:
#line 1513 "go.y"
    {
		(yyval.node) = nod(OTINTER, N, N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
		fixlbrace((yyvsp[(2) - (5)].i));
	}
    break;

  case 227:
#line 1519 "go.y"
    {
		(yyval.node) = nod(OTINTER, N, N);
		fixlbrace((yyvsp[(2) - (3)].i));
	}
    break;

  case 228:
#line 1530 "go.y"
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

  case 229:
#line 1546 "go.y"
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

  case 230:
#line 1575 "go.y"
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

  case 231:
#line 1613 "go.y"
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

  case 232:
#line 1638 "go.y"
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

  case 233:
#line 1656 "go.y"
    {
		(yyvsp[(3) - (5)].list) = checkarglist((yyvsp[(3) - (5)].list), 1);
		(yyval.node) = nod(OTFUNC, N, N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
		(yyval.node)->rlist = (yyvsp[(5) - (5)].list);
	}
    break;

  case 234:
#line 1664 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 235:
#line 1668 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (3)].list);
		if((yyval.list) == nil)
			(yyval.list) = list1(nod(OEMPTY, N, N));
	}
    break;

  case 236:
#line 1676 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 237:
#line 1680 "go.y"
    {
		(yyval.list) = list1(nod(ODCLFIELD, N, (yyvsp[(1) - (1)].node)));
	}
    break;

  case 238:
#line 1684 "go.y"
    {
		(yyvsp[(2) - (3)].list) = checkarglist((yyvsp[(2) - (3)].list), 0);
		(yyval.list) = (yyvsp[(2) - (3)].list);
	}
    break;

  case 239:
#line 1691 "go.y"
    {
		closurehdr((yyvsp[(1) - (1)].node));
	}
    break;

  case 240:
#line 1697 "go.y"
    {
		(yyval.node) = closurebody((yyvsp[(3) - (4)].list));
		fixlbrace((yyvsp[(2) - (4)].i));
	}
    break;

  case 241:
#line 1702 "go.y"
    {
		(yyval.node) = closurebody(nil);
	}
    break;

  case 242:
#line 1713 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 243:
#line 1717 "go.y"
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

  case 245:
#line 1730 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list));
	}
    break;

  case 247:
#line 1737 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list));
	}
    break;

  case 248:
#line 1743 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 249:
#line 1747 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 251:
#line 1754 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list));
	}
    break;

  case 252:
#line 1760 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 253:
#line 1764 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 254:
#line 1770 "go.y"
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

  case 255:
#line 1793 "go.y"
    {
		(yyvsp[(1) - (2)].node)->val = (yyvsp[(2) - (2)].val);
		(yyval.list) = list1((yyvsp[(1) - (2)].node));
	}
    break;

  case 256:
#line 1798 "go.y"
    {
		(yyvsp[(2) - (4)].node)->val = (yyvsp[(4) - (4)].val);
		(yyval.list) = list1((yyvsp[(2) - (4)].node));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 257:
#line 1804 "go.y"
    {
		(yyvsp[(2) - (3)].node)->right = nod(OIND, (yyvsp[(2) - (3)].node)->right, N);
		(yyvsp[(2) - (3)].node)->val = (yyvsp[(3) - (3)].val);
		(yyval.list) = list1((yyvsp[(2) - (3)].node));
	}
    break;

  case 258:
#line 1810 "go.y"
    {
		(yyvsp[(3) - (5)].node)->right = nod(OIND, (yyvsp[(3) - (5)].node)->right, N);
		(yyvsp[(3) - (5)].node)->val = (yyvsp[(5) - (5)].val);
		(yyval.list) = list1((yyvsp[(3) - (5)].node));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 259:
#line 1817 "go.y"
    {
		(yyvsp[(3) - (5)].node)->right = nod(OIND, (yyvsp[(3) - (5)].node)->right, N);
		(yyvsp[(3) - (5)].node)->val = (yyvsp[(5) - (5)].val);
		(yyval.list) = list1((yyvsp[(3) - (5)].node));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 260:
#line 1826 "go.y"
    {
		Node *n;

		(yyval.sym) = (yyvsp[(1) - (1)].sym);
		n = oldname((yyvsp[(1) - (1)].sym));
		if(n->pack != N)
			n->pack->used = 1;
	}
    break;

  case 261:
#line 1835 "go.y"
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

  case 262:
#line 1850 "go.y"
    {
		(yyval.node) = embedded((yyvsp[(1) - (1)].sym), localpkg);
	}
    break;

  case 263:
#line 1856 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, (yyvsp[(1) - (2)].node), (yyvsp[(2) - (2)].node));
		ifacedcl((yyval.node));
	}
    break;

  case 264:
#line 1861 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, oldname((yyvsp[(1) - (1)].sym)));
	}
    break;

  case 265:
#line 1865 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, oldname((yyvsp[(2) - (3)].sym)));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 266:
#line 1872 "go.y"
    {
		// without func keyword
		(yyvsp[(2) - (4)].list) = checkarglist((yyvsp[(2) - (4)].list), 1);
		(yyval.node) = nod(OTFUNC, fakethis(), N);
		(yyval.node)->list = (yyvsp[(2) - (4)].list);
		(yyval.node)->rlist = (yyvsp[(4) - (4)].list);
	}
    break;

  case 268:
#line 1886 "go.y"
    {
		(yyval.node) = nod(ONONAME, N, N);
		(yyval.node)->sym = (yyvsp[(1) - (2)].sym);
		(yyval.node) = nod(OKEY, (yyval.node), (yyvsp[(2) - (2)].node));
	}
    break;

  case 269:
#line 1892 "go.y"
    {
		(yyval.node) = nod(ONONAME, N, N);
		(yyval.node)->sym = (yyvsp[(1) - (2)].sym);
		(yyval.node) = nod(OKEY, (yyval.node), (yyvsp[(2) - (2)].node));
	}
    break;

  case 271:
#line 1901 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 272:
#line 1905 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 273:
#line 1910 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 274:
#line 1914 "go.y"
    {
		(yyval.list) = (yyvsp[(1) - (2)].list);
	}
    break;

  case 275:
#line 1922 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 277:
#line 1927 "go.y"
    {
		(yyval.node) = liststmt((yyvsp[(1) - (1)].list));
	}
    break;

  case 279:
#line 1932 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 285:
#line 1943 "go.y"
    {
		(yyvsp[(1) - (2)].node) = nod(OLABEL, (yyvsp[(1) - (2)].node), N);
		(yyvsp[(1) - (2)].node)->sym = dclstack;  // context, for goto restrictions
	}
    break;

  case 286:
#line 1948 "go.y"
    {
		NodeList *l;

		(yyvsp[(1) - (4)].node)->defn = (yyvsp[(4) - (4)].node);
		l = list1((yyvsp[(1) - (4)].node));
		if((yyvsp[(4) - (4)].node))
			l = list(l, (yyvsp[(4) - (4)].node));
		(yyval.node) = liststmt(l);
	}
    break;

  case 287:
#line 1958 "go.y"
    {
		// will be converted to OFALL
		(yyval.node) = nod(OXFALL, N, N);
		(yyval.node)->xoffset = block;
	}
    break;

  case 288:
#line 1964 "go.y"
    {
		(yyval.node) = nod(OBREAK, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 289:
#line 1968 "go.y"
    {
		(yyval.node) = nod(OCONTINUE, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 290:
#line 1972 "go.y"
    {
		(yyval.node) = nod(OPROC, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 291:
#line 1976 "go.y"
    {
		(yyval.node) = nod(ODEFER, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 292:
#line 1980 "go.y"
    {
		(yyval.node) = nod(OGOTO, (yyvsp[(2) - (2)].node), N);
		(yyval.node)->sym = dclstack;  // context, for goto restrictions
	}
    break;

  case 293:
#line 1985 "go.y"
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

  case 294:
#line 2004 "go.y"
    {
		(yyval.list) = nil;
		if((yyvsp[(1) - (1)].node) != N)
			(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 295:
#line 2010 "go.y"
    {
		(yyval.list) = (yyvsp[(1) - (3)].list);
		if((yyvsp[(3) - (3)].node) != N)
			(yyval.list) = list((yyval.list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 296:
#line 2018 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 297:
#line 2022 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 298:
#line 2028 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 299:
#line 2032 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 300:
#line 2038 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 301:
#line 2042 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 302:
#line 2048 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 303:
#line 2052 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 304:
#line 2061 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 305:
#line 2065 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 306:
#line 2069 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 307:
#line 2073 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 308:
#line 2078 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 309:
#line 2082 "go.y"
    {
		(yyval.list) = (yyvsp[(1) - (2)].list);
	}
    break;

  case 314:
#line 2096 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 316:
#line 2102 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 318:
#line 2108 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 320:
#line 2114 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 322:
#line 2120 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 324:
#line 2126 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 326:
#line 2132 "go.y"
    {
		(yyval.val).ctype = CTxxx;
	}
    break;

  case 328:
#line 2142 "go.y"
    {
		importimport((yyvsp[(2) - (4)].sym), (yyvsp[(3) - (4)].val).u.sval);
	}
    break;

  case 329:
#line 2146 "go.y"
    {
		importvar((yyvsp[(2) - (4)].sym), (yyvsp[(3) - (4)].type));
	}
    break;

  case 330:
#line 2150 "go.y"
    {
		importconst((yyvsp[(2) - (5)].sym), types[TIDEAL], (yyvsp[(4) - (5)].node));
	}
    break;

  case 331:
#line 2154 "go.y"
    {
		importconst((yyvsp[(2) - (6)].sym), (yyvsp[(3) - (6)].type), (yyvsp[(5) - (6)].node));
	}
    break;

  case 332:
#line 2158 "go.y"
    {
		importtype((yyvsp[(2) - (4)].type), (yyvsp[(3) - (4)].type));
	}
    break;

  case 333:
#line 2162 "go.y"
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

  case 334:
#line 2182 "go.y"
    {
		(yyval.sym) = (yyvsp[(1) - (1)].sym);
		structpkg = (yyval.sym)->pkg;
	}
    break;

  case 335:
#line 2189 "go.y"
    {
		(yyval.type) = pkgtype((yyvsp[(1) - (1)].sym));
		importsym((yyvsp[(1) - (1)].sym), OTYPE);
	}
    break;

  case 341:
#line 2209 "go.y"
    {
		(yyval.type) = pkgtype((yyvsp[(1) - (1)].sym));
	}
    break;

  case 342:
#line 2213 "go.y"
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

  case 343:
#line 2223 "go.y"
    {
		(yyval.type) = aindex(N, (yyvsp[(3) - (3)].type));
	}
    break;

  case 344:
#line 2227 "go.y"
    {
		(yyval.type) = aindex(nodlit((yyvsp[(2) - (4)].val)), (yyvsp[(4) - (4)].type));
	}
    break;

  case 345:
#line 2231 "go.y"
    {
		(yyval.type) = maptype((yyvsp[(3) - (5)].type), (yyvsp[(5) - (5)].type));
	}
    break;

  case 346:
#line 2235 "go.y"
    {
		(yyval.type) = tostruct((yyvsp[(3) - (4)].list));
	}
    break;

  case 347:
#line 2239 "go.y"
    {
		(yyval.type) = tointerface((yyvsp[(3) - (4)].list));
	}
    break;

  case 348:
#line 2243 "go.y"
    {
		(yyval.type) = ptrto((yyvsp[(2) - (2)].type));
	}
    break;

  case 349:
#line 2247 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(2) - (2)].type);
		(yyval.type)->chan = Cboth;
	}
    break;

  case 350:
#line 2253 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(3) - (4)].type);
		(yyval.type)->chan = Cboth;
	}
    break;

  case 351:
#line 2259 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(3) - (3)].type);
		(yyval.type)->chan = Csend;
	}
    break;

  case 352:
#line 2267 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(3) - (3)].type);
		(yyval.type)->chan = Crecv;
	}
    break;

  case 353:
#line 2275 "go.y"
    {
		(yyval.type) = functype(nil, (yyvsp[(3) - (5)].list), (yyvsp[(5) - (5)].list));
	}
    break;

  case 354:
#line 2281 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, typenod((yyvsp[(2) - (3)].type)));
		if((yyvsp[(1) - (3)].sym))
			(yyval.node)->left = newname((yyvsp[(1) - (3)].sym));
		(yyval.node)->val = (yyvsp[(3) - (3)].val);
	}
    break;

  case 355:
#line 2288 "go.y"
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

  case 356:
#line 2304 "go.y"
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

  case 357:
#line 2326 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, newname((yyvsp[(1) - (5)].sym)), typenod(functype(fakethis(), (yyvsp[(3) - (5)].list), (yyvsp[(5) - (5)].list))));
	}
    break;

  case 358:
#line 2330 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, typenod((yyvsp[(1) - (1)].type)));
	}
    break;

  case 359:
#line 2335 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 361:
#line 2342 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (3)].list);
	}
    break;

  case 362:
#line 2346 "go.y"
    {
		(yyval.list) = list1(nod(ODCLFIELD, N, typenod((yyvsp[(1) - (1)].type))));
	}
    break;

  case 363:
#line 2356 "go.y"
    {
		(yyval.node) = nodlit((yyvsp[(1) - (1)].val));
	}
    break;

  case 364:
#line 2360 "go.y"
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

  case 365:
#line 2379 "go.y"
    {
		(yyval.node) = oldname(pkglookup((yyvsp[(1) - (1)].sym)->name, builtinpkg));
		if((yyval.node)->op != OLITERAL)
			yyerror("bad constant %S", (yyval.node)->sym);
	}
    break;

  case 367:
#line 2388 "go.y"
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

  case 370:
#line 2404 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 371:
#line 2408 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 372:
#line 2414 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 373:
#line 2418 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 374:
#line 2424 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 375:
#line 2428 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;


/* Line 1267 of yacc.c.  */
#line 5215 "y.tab.c"
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


#line 2432 "go.y"


static void
fixlbrace(int lbr)
{
	// If the opening brace was an LBODY,
	// set up for another one now that we're done.
	// See comment in lex.c about loophack.
	if(lbr == LBODY)
		loophack = 1;
}


